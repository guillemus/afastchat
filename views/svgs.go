package views

import (
	. "maragu.dev/gomponents"
)

func iconPlus() Node {
	return El("svg",
		Attr("xmlns", "http://www.w3.org/2000/svg"),
		Attr("viewBox", "0 0 24 24"),
		Attr("fill", "none"),
		Attr("stroke", "currentColor"),
		Attr("stroke-width", "2.5"),
		Attr("stroke-linecap", "round"),
		El("line", Attr("x1", "12"), Attr("y1", "5"), Attr("x2", "12"), Attr("y2", "19")),
		El("line", Attr("x1", "5"), Attr("y1", "12"), Attr("x2", "19"), Attr("y2", "12")),
	)
}

func iconSearch() Node {
	return El("svg",
		Attr("xmlns", "http://www.w3.org/2000/svg"),
		Attr("viewBox", "0 0 24 24"),
		Attr("fill", "none"),
		Attr("stroke", "currentColor"),
		Attr("stroke-width", "2"),
		El("circle", Attr("cx", "11"), Attr("cy", "11"), Attr("r", "8")),
		El("path", Attr("d", "m21 21-4.3-4.3")),
	)
}

func iconFolder() Node {
	return El("svg",
		Attr("xmlns", "http://www.w3.org/2000/svg"),
		Attr("viewBox", "0 0 24 24"),
		Attr("fill", "none"),
		Attr("stroke", "currentColor"),
		Attr("stroke-width", "2"),
		Attr("stroke-linecap", "round"),
		Attr("stroke-linejoin", "round"),
		El("path", Attr("d", "M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2Z")),
	)
}

func iconFolderOpen() Node {
	return El("svg",
		Attr("xmlns", "http://www.w3.org/2000/svg"),
		Attr("viewBox", "0 0 24 24"),
		Attr("fill", "none"),
		Attr("stroke", "currentColor"),
		Attr("stroke-width", "2"),
		Attr("stroke-linecap", "round"),
		Attr("stroke-linejoin", "round"),
		El("path", Attr("d", "M3 7a2 2 0 0 1 2-2h4l2 2h7a2 2 0 0 1 2 2v1")),
		El("path", Attr("d", "M3 17.5 5.2 10h17l-2.1 7.5A2 2 0 0 1 18.2 19H5a2 2 0 0 1-2-1.5Z")),
	)
}

func iconFolderPlus() Node {
	return El("svg",
		Attr("xmlns", "http://www.w3.org/2000/svg"),
		Attr("viewBox", "0 0 24 24"),
		Attr("fill", "none"),
		Attr("stroke", "currentColor"),
		Attr("stroke-width", "2"),
		Attr("stroke-linecap", "round"),
		Attr("stroke-linejoin", "round"),
		El("path", Attr("d", "M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2Z")),
		El("path", Attr("d", "M9 13h6")),
		El("path", Attr("d", "M12 10v6")),
	)
}

func iconChat() Node {
	return El("svg",
		Attr("xmlns", "http://www.w3.org/2000/svg"),
		Attr("viewBox", "0 0 24 24"),
		Attr("fill", "none"),
		Attr("stroke", "currentColor"),
		Attr("stroke-width", "2"),
		Attr("stroke-linecap", "round"),
		Attr("stroke-linejoin", "round"),
		El("path", Attr("d", "M21 15a4 4 0 0 1-4 4H8l-5 3V7a4 4 0 0 1 4-4h10a4 4 0 0 1 4 4Z")),
	)
}

func iconArchive() Node {
	return El("svg",
		Attr("xmlns", "http://www.w3.org/2000/svg"),
		Attr("viewBox", "0 0 24 24"),
		Attr("fill", "none"),
		Attr("stroke", "currentColor"),
		Attr("stroke-width", "2"),
		Attr("stroke-linecap", "round"),
		Attr("stroke-linejoin", "round"),
		El("polyline", Attr("points", "3 6 5 6 21 6")),
		El("path", Attr("d", "M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6")),
		El("line", Attr("x1", "10"), Attr("y1", "11"), Attr("x2", "10"), Attr("y2", "17")),
		El("line", Attr("x1", "14"), Attr("y1", "11"), Attr("x2", "14"), Attr("y2", "17")),
	)
}

func iconSpinner() Node {
	return El("svg",
		Attr("xmlns", "http://www.w3.org/2000/svg"),
		Attr("viewBox", "0 0 24 24"),
		Attr("fill", "none"),
		Attr("stroke", "currentColor"),
		Attr("stroke-width", "2.5"),
		El("circle",
			Attr("cx", "12"),
			Attr("cy", "12"),
			Attr("r", "9"),
			Attr("stroke-opacity", "0.25"),
		),
		El("path",
			Attr("d", "M12 3a9 9 0 0 1 9 9"),
			Attr("stroke-linecap", "round"),
		),
	)
}

func iconArrowUp() Node {
	return El("svg",
		Attr("xmlns", "http://www.w3.org/2000/svg"),
		Attr("viewBox", "0 0 24 24"),
		Attr("fill", "none"),
		Attr("stroke", "currentColor"),
		Attr("stroke-width", "2.5"),
		Attr("stroke-linecap", "round"),
		Attr("stroke-linejoin", "round"),
		El("path", Attr("d", "M12 19V5")),
		El("path", Attr("d", "M5 12l7-7 7 7")),
	)
}

func iconArrowDown() Node {
	return El("svg",
		Attr("xmlns", "http://www.w3.org/2000/svg"),
		Attr("viewBox", "0 0 24 24"),
		Attr("fill", "none"),
		Attr("stroke", "currentColor"),
		Attr("stroke-width", "2.25"),
		Attr("stroke-linecap", "round"),
		Attr("stroke-linejoin", "round"),
		El("path", Attr("d", "M12 5v14")),
		El("path", Attr("d", "m6 13 6 6 6-6")),
	)
}

func iconSquare() Node {
	return El("svg",
		Attr("xmlns", "http://www.w3.org/2000/svg"),
		Attr("viewBox", "0 0 24 24"),
		Attr("fill", "currentColor"),
		El("rect", Attr("x", "5"), Attr("y", "5"), Attr("width", "14"), Attr("height", "14"), Attr("rx", "2")),
	)
}
