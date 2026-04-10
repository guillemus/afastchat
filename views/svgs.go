package views

import g "maragu.dev/gomponents"

func iconPlus() g.Node {
	return g.El("svg",
		g.Attr("xmlns", "http://www.w3.org/2000/svg"),
		g.Attr("viewBox", "0 0 24 24"),
		g.Attr("fill", "none"),
		g.Attr("stroke", "currentColor"),
		g.Attr("stroke-width", "2.5"),
		g.Attr("stroke-linecap", "round"),
		g.El("line", g.Attr("x1", "12"), g.Attr("y1", "5"), g.Attr("x2", "12"), g.Attr("y2", "19")),
		g.El("line", g.Attr("x1", "5"), g.Attr("y1", "12"), g.Attr("x2", "19"), g.Attr("y2", "12")),
	)
}

func iconSearch() g.Node {
	return g.El("svg",
		g.Attr("xmlns", "http://www.w3.org/2000/svg"),
		g.Attr("viewBox", "0 0 24 24"),
		g.Attr("fill", "none"),
		g.Attr("stroke", "currentColor"),
		g.Attr("stroke-width", "2"),
		g.El("circle", g.Attr("cx", "11"), g.Attr("cy", "11"), g.Attr("r", "8")),
		g.El("path", g.Attr("d", "m21 21-4.3-4.3")),
	)
}

func iconArchive() g.Node {
	return g.El("svg",
		g.Attr("xmlns", "http://www.w3.org/2000/svg"),
		g.Attr("viewBox", "0 0 24 24"),
		g.Attr("fill", "none"),
		g.Attr("stroke", "currentColor"),
		g.Attr("stroke-width", "2"),
		g.Attr("stroke-linecap", "round"),
		g.Attr("stroke-linejoin", "round"),
		g.El("polyline", g.Attr("points", "3 6 5 6 21 6")),
		g.El("path", g.Attr("d", "M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6")),
		g.El("line", g.Attr("x1", "10"), g.Attr("y1", "11"), g.Attr("x2", "10"), g.Attr("y2", "17")),
		g.El("line", g.Attr("x1", "14"), g.Attr("y1", "11"), g.Attr("x2", "14"), g.Attr("y2", "17")),
	)
}

func iconArrowUp() g.Node {
	return g.El("svg",
		g.Attr("xmlns", "http://www.w3.org/2000/svg"),
		g.Attr("viewBox", "0 0 24 24"),
		g.Attr("fill", "none"),
		g.Attr("stroke", "currentColor"),
		g.Attr("stroke-width", "2.5"),
		g.Attr("stroke-linecap", "round"),
		g.Attr("stroke-linejoin", "round"),
		g.El("path", g.Attr("d", "M12 19V5")),
		g.El("path", g.Attr("d", "M5 12l7-7 7 7")),
	)
}

func iconSquare() g.Node {
	return g.El("svg",
		g.Attr("xmlns", "http://www.w3.org/2000/svg"),
		g.Attr("viewBox", "0 0 24 24"),
		g.Attr("fill", "currentColor"),
		g.El("rect", g.Attr("x", "5"), g.Attr("y", "5"), g.Attr("width", "14"), g.Attr("height", "14"), g.Attr("rx", "2")),
	)
}
