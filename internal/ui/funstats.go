package ui

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func FunStatsPage(d FunStatsData) []g.Node {
	return []g.Node{
		pageTitle("Fun Stats"),
		P(
			Class("mt-4 text-xl font-medium text-zinc-800 dark:text-zinc-200"),
			g.Text("Users and ads over time"),
		),
		Div(
			Class("mt-10 space-y-8"),
			funStatsChartCard(
				"Users",
				"Monthly totals of registered users and users "+
					"with at least one active ad.",
				d.Months,
				[]funStatsSeries{
					{
						Name:  "Registered users",
						Color: "#3b82f6",
						Values: funStatsValues(d.Months,
							func(m FunStatsMonth) int {
								return m.RegisteredUsers
							}),
					},
					{
						Name:  "Users with active ads",
						Color: "#22c55e",
						Values: funStatsValues(d.Months,
							func(m FunStatsMonth) int {
								return m.UsersWithActiveAds
							}),
					},
				},
			),
			funStatsChartCard(
				"Active ads",
				"Monthly total of ads that were active.",
				d.Months,
				[]funStatsSeries{
					{
						Name:  "Active ads",
						Color: "#a855f7",
						Values: funStatsValues(d.Months,
							func(m FunStatsMonth) int {
								return m.ActiveAds
							}),
					},
				},
			),
		),
	}
}

type funStatsSeries struct {
	Name   string
	Color  string
	Values []int
}

func funStatsValues(months []FunStatsMonth,
	value func(FunStatsMonth) int) []int {
	out := make([]int, len(months))
	for i, m := range months {
		out[i] = value(m)
	}
	return out
}

func funStatsChartCard(title, subtitle string, months []FunStatsMonth,
	series []funStatsSeries) g.Node {
	return Div(
		Class("p-4 border border-zinc-200 dark:border-zinc-700 "+
			"rounded-lg space-y-4"),
		H2(
			Class("text-lg font-semibold text-zinc-900 dark:text-zinc-100"),
			g.Text(title),
		),
		P(
			Class("text-sm text-zinc-600 dark:text-zinc-400"),
			g.Text(subtitle),
		),
		funStatsLegend(series),
		funStatsLineChart(months, series),
	)
}

func funStatsLegend(series []funStatsSeries) g.Node {
	items := make([]g.Node, len(series))
	for i, s := range series {
		items[i] = Div(
			Class("flex items-center gap-2 text-sm text-zinc-700 "+
				"dark:text-zinc-300"),
			Span(
				Class("inline-block w-3 h-3 rounded-sm shrink-0"),
				Style("background-color: "+s.Color),
			),
			g.Text(s.Name),
		)
	}
	return Div(Class("flex flex-wrap gap-x-4 gap-y-2"), g.Group(items))
}

func funStatsLineChart(months []FunStatsMonth,
	series []funStatsSeries) g.Node {
	const (
		width  = 640.0
		height = 280.0
		left   = 48.0
		right  = 16.0
		top    = 16.0
		bottom = 40.0
	)
	plotW := width - left - right
	plotH := height - top - bottom
	plotBottom := top + plotH

	maxVal := 0
	for _, s := range series {
		for _, v := range s.Values {
			if v > maxVal {
				maxVal = v
			}
		}
	}
	topVal, ticks := funStatsNiceScale(maxVal)
	labels := make([]string, len(months))
	for i, m := range months {
		labels[i] = m.Label
	}

	children := []g.Node{
		g.Attr("xmlns", "http://www.w3.org/2000/svg"),
		g.Attr("viewBox", fmt.Sprintf("0 0 %.0f %.0f", width, height)),
		g.Attr("preserveAspectRatio", "xMidYMid meet"),
		g.Attr("role", "img"),
		g.Attr("aria-label", "Monthly line chart"),
		Class("w-full h-auto text-zinc-500 dark:text-zinc-400"),
	}
	children = append(children,
		funStatsGrid(left, top, plotW, plotH, plotBottom, topVal,
			ticks)...)
	children = append(children,
		funStatsXLabels(labels, left, plotW, plotBottom)...)
	for _, s := range series {
		children = append(children,
			funStatsPolyline(s, months, left, plotW, plotBottom, plotH,
				topVal)...)
	}
	return g.El("svg", children...)
}

func funStatsGrid(left, top, plotW, plotH, plotBottom float64,
	topVal int, ticks []int) []g.Node {
	nodes := []g.Node{
		g.El("line",
			g.Attr("x1", fmt.Sprintf("%.1f", left)),
			g.Attr("y1", fmt.Sprintf("%.1f", top)),
			g.Attr("x2", fmt.Sprintf("%.1f", left)),
			g.Attr("y2", fmt.Sprintf("%.1f", plotBottom)),
			g.Attr("stroke", "currentColor"),
			g.Attr("stroke-width", "1"),
		),
		g.El("line",
			g.Attr("x1", fmt.Sprintf("%.1f", left)),
			g.Attr("y1", fmt.Sprintf("%.1f", plotBottom)),
			g.Attr("x2", fmt.Sprintf("%.1f", left+plotW)),
			g.Attr("y2", fmt.Sprintf("%.1f", plotBottom)),
			g.Attr("stroke", "currentColor"),
			g.Attr("stroke-width", "1"),
		),
	}
	for _, tick := range ticks {
		y := funStatsY(tick, topVal, plotBottom, plotH)
		nodes = append(nodes,
			g.El("line",
				g.Attr("x1", fmt.Sprintf("%.1f", left)),
				g.Attr("y1", fmt.Sprintf("%.1f", y)),
				g.Attr("x2", fmt.Sprintf("%.1f", left+plotW)),
				g.Attr("y2", fmt.Sprintf("%.1f", y)),
				g.Attr("stroke", "currentColor"),
				g.Attr("stroke-width", "1"),
				g.Attr("stroke-opacity", "0.25"),
			),
			g.El("text",
				g.Attr("x", fmt.Sprintf("%.1f", left-8)),
				g.Attr("y", fmt.Sprintf("%.1f", y+4)),
				g.Attr("text-anchor", "end"),
				g.Attr("font-size", "11"),
				g.Attr("fill", "currentColor"),
				g.Text(strconv.Itoa(tick)),
			),
		)
	}
	return nodes
}

func funStatsXLabels(labels []string, left, plotW,
	plotBottom float64) []g.Node {
	n := len(labels)
	if n == 0 {
		return nil
	}
	step := funStatsLabelStep(n)
	var nodes []g.Node
	for i, label := range labels {
		if i%step != 0 && i != n-1 {
			continue
		}
		x := funStatsX(i, n, left, plotW)
		nodes = append(nodes,
			g.El("text",
				g.Attr("x", fmt.Sprintf("%.1f", x)),
				g.Attr("y", fmt.Sprintf("%.1f", plotBottom+18)),
				g.Attr("text-anchor", "middle"),
				g.Attr("font-size", "11"),
				g.Attr("fill", "currentColor"),
				g.Text(label),
			),
		)
	}
	return nodes
}

func funStatsPolyline(s funStatsSeries, months []FunStatsMonth,
	left, plotW, plotBottom, plotH float64, topVal int) []g.Node {
	n := len(s.Values)
	if n == 0 {
		return nil
	}
	points := make([]string, n)
	var nodes []g.Node
	for i, v := range s.Values {
		x := funStatsX(i, n, left, plotW)
		y := funStatsY(v, topVal, plotBottom, plotH)
		points[i] = fmt.Sprintf("%.1f,%.1f", x, y)
		label := ""
		if i < len(months) {
			label = months[i].Label
		}
		nodes = append(nodes,
			g.El("circle",
				g.Attr("cx", fmt.Sprintf("%.1f", x)),
				g.Attr("cy", fmt.Sprintf("%.1f", y)),
				g.Attr("r", "3.5"),
				g.Attr("fill", s.Color),
				g.El("title",
					g.Textf("%s: %s %s", label, s.Name,
						strconv.Itoa(v))),
			),
		)
	}
	line := g.El("polyline",
		g.Attr("fill", "none"),
		g.Attr("stroke", s.Color),
		g.Attr("stroke-width", "2"),
		g.Attr("stroke-linejoin", "round"),
		g.Attr("stroke-linecap", "round"),
		g.Attr("points", strings.Join(points, " ")),
	)
	return append([]g.Node{line}, nodes...)
}

func funStatsX(i, n int, left, plotW float64) float64 {
	if n <= 1 {
		return left + plotW/2
	}
	return left + float64(i)*plotW/float64(n-1)
}

func funStatsY(val, topVal int, plotBottom, plotH float64) float64 {
	if topVal <= 0 {
		return plotBottom
	}
	return plotBottom - (float64(val)/float64(topVal))*plotH
}

func funStatsLabelStep(n int) int {
	switch {
	case n <= 12:
		return 1
	case n <= 24:
		return 2
	case n <= 48:
		return 3
	default:
		return 6
	}
}

func funStatsNiceScale(maxVal int) (int, []int) {
	if maxVal <= 0 {
		return 1, []int{0, 1}
	}
	span := funStatsNiceNum(float64(maxVal), false)
	step := funStatsNiceNum(span/4, true)
	if step < 1 {
		step = 1
	}
	top := int(math.Ceil(float64(maxVal)/step) * step)
	var ticks []int
	for v := 0.0; v <= float64(top)+step/2; v += step {
		n := int(math.Round(v))
		if n > top {
			break
		}
		ticks = append(ticks, n)
	}
	if ticks[len(ticks)-1] != top {
		ticks = append(ticks, top)
	}
	return top, ticks
}

func funStatsNiceNum(x float64, round bool) float64 {
	if x <= 0 {
		return 1
	}
	exp := math.Floor(math.Log10(x))
	f := x / math.Pow(10, exp)
	var nf float64
	if round {
		switch {
		case f < 1.5:
			nf = 1
		case f < 3:
			nf = 2
		case f < 7:
			nf = 5
		default:
			nf = 10
		}
	} else {
		switch {
		case f <= 1:
			nf = 1
		case f <= 2:
			nf = 2
		case f <= 5:
			nf = 5
		default:
			nf = 10
		}
	}
	return nf * math.Pow(10, exp)
}
