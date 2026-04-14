package db

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/klauspost/compress/zstd"
	"zombiezen.com/go/sqlite"
)

var messageContentMagic = []byte("afastchat:zstd\x00")

var messageEncoderPool = sync.Pool{
	New: func() any {
		encoder, err := zstd.NewWriter(nil)
		if err != nil {
			panic(fmt.Sprintf("create zstd encoder: %v", err))
		}
		return encoder
	},
}

var messageDecoderPool = sync.Pool{
	New: func() any {
		decoder, err := zstd.NewReader(nil)
		if err != nil {
			panic(fmt.Sprintf("create zstd decoder: %v", err))
		}
		return decoder
	},
}

func compressMessageContent(content string) []byte {
	src := []byte(content)
	encoder := messageEncoderPool.Get().(*zstd.Encoder)
	defer messageEncoderPool.Put(encoder)

	dst := append([]byte{}, messageContentMagic...)
	return encoder.EncodeAll(src, dst)
}

func decompressMessageContent(content []byte) (string, error) {
	if !bytes.HasPrefix(content, messageContentMagic) {
		return "", fmt.Errorf("missing zstd message header")
	}

	decoder := messageDecoderPool.Get().(*zstd.Decoder)
	defer messageDecoderPool.Put(decoder)

	decoded, err := decoder.DecodeAll(content[len(messageContentMagic):], nil)
	if err != nil {
		return "", fmt.Errorf("decode zstd message: %w", err)
	}

	return string(decoded), nil
}

func scanMessageContent(stmt *sqlite.Stmt) (string, error) {
	contentType := stmt.ColumnType(stmt.ColumnIndex("content"))
	if contentType != sqlite.TypeBlob {
		return "", fmt.Errorf("unexpected content type: %s", contentType)
	}

	content := make([]byte, stmt.GetLen("content"))
	stmt.GetBytes("content", content)

	decoded, err := decompressMessageContent(content)
	if err != nil {
		return "", fmt.Errorf("read message content: %w", err)
	}

	return decoded, nil
}
