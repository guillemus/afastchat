package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"afastchat.com/db"
)

const (
	attachmentObjectPrefix = "attachments/"
	maxImageCount          = 4
	maxImageBytes          = 5 << 20
	maxSendBodyBytes       = maxImageCount*maxImageBytes*4/3 + 1<<20
)

type imageSignal struct {
	Name     string `json:"name"`
	Contents string `json:"contents"`
	Mime     string `json:"mime"`
}

type uploadedImage struct {
	filename string
	mimeType string
	data     []byte
}

func uploadedImagesFromSignals(files []imageSignal) ([]uploadedImage, error) {
	if len(files) > maxImageCount {
		return nil, fmt.Errorf("Attach at most %d images.", maxImageCount)
	}

	images := make([]uploadedImage, 0, len(files))
	for _, file := range files {
		data, err := base64.StdEncoding.Strict().DecodeString(file.Contents)
		if err != nil {
			return nil, errors.New("An image is invalid. Remove it and try again.")
		}
		if len(data) < 1 || len(data) > maxImageBytes {
			return nil, fmt.Errorf("Keep each image under %d MB.", maxImageBytes>>20)
		}

		mimeType := http.DetectContentType(data)
		if imageExtension(mimeType) == "" {
			return nil, errors.New("Only PNG, JPEG, GIF, and WebP images are supported.")
		}

		images = append(images, uploadedImage{
			filename: imageFileName(filepath.Base(file.Name), mimeType),
			mimeType: mimeType,
			data:     data,
		})
	}

	return images, nil
}

func imageExtension(mimeType string) string {
	switch mimeType {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}

func imageFileName(filename, mimeType string) string {
	stem := strings.TrimSpace(strings.TrimSuffix(filename, filepath.Ext(filename)))
	if stem == "" || stem == "." || stem == string(filepath.Separator) {
		stem = "image"
	}
	return stem + imageExtension(mimeType)
}

func newObjectKey(prefix string) (string, error) {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return "", fmt.Errorf("generate object key: %w", err)
	}
	return prefix + hex.EncodeToString(id), nil
}

func (x *Server) storeAttachments(ctx context.Context, userID string, uploads []uploadedImage) ([]db.AttachmentInput, error) {
	attachments := make([]db.AttachmentInput, 0, len(uploads))
	for _, upload := range uploads {
		key, err := newObjectKey(attachmentObjectPrefix + userID + "/")
		if err != nil {
			return nil, err
		}
		if err := x.r2Store.Put(ctx, key, upload.mimeType, upload.data); err != nil {
			return nil, err
		}
		attachments = append(attachments, db.AttachmentInput{
			ObjectKey: key, MimeType: upload.mimeType, Filename: upload.filename,
		})
	}
	return attachments, nil
}

func (x *Server) deleteAttachmentObjects(objectKeys []string) {
	if len(objectKeys) == 0 {
		return
	}

	GoRecover(func() {
		ctx := context.Background()
		for _, key := range objectKeys {
			if err := x.r2Store.Delete(ctx, key); err != nil {
				slog.Error("delete attachment object", "object_key", key, "err", err)
			}
		}
	})
}

func (x *Server) getThreadMessages(ctx context.Context, userID, threadID string) ([]db.Message, error) {
	messages, err := x.db.GetMessagesByThread(ctx, userID, threadID)
	if err != nil {
		return nil, err
	}

	attachments, err := x.db.ListThreadAttachments(ctx, userID, threadID)
	if err != nil {
		return nil, err
	}

	for i, message := range messages {
		for _, attachment := range attachments[message.ID] {
			attachment.URL, err = x.r2Store.PresignGet(ctx, attachment.ObjectKey)
			if err != nil {
				return nil, err
			}
			messages[i].Attachments = append(messages[i].Attachments, attachment)
		}
	}
	return messages, nil
}
