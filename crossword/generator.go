package crossword

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Dir byte

const (
	Horizontal Dir = iota // ---
	Vertical              // |
)

type Point struct {
	x int
	y int
}

func (p Point) String() string {
	return "(" + strconv.Itoa(p.x) + ", " + strconv.Itoa(p.y) + ")"
}

type Char struct {
	root         Point
	p            Point
	c            string
	d            Dir
	master       string
	intersection bool
}

func (c Char) String() string {
	return "char{" + c.root.String() + ", " + c.p.String() + ", " + c.c + ", " + strconv.Itoa(int(c.d)) + "," + c.master + "," + strconv.FormatBool(c.intersection) + "}"
}

type String struct {
	root Point
	s    string
	d    Dir
}

func (s String) String() string {
	return "string{" + s.root.String() + ", " + s.s + ", " + strconv.Itoa(int(s.d)) + "}"
}

type Intersection struct {
	root   Point
	points []Point
	dir    Dir
	length int
}

// ------> x, width
//|
//|
//v y, height
type Generator struct {
	locationOfChars   map[Point]Char
	locationOfStrings map[Point]String
	stringOfLocations map[String]Point
	width             int
	height            int
	desireSize        int
	maxWidth          int
	maxHeight         int
}

func NewGenerator(desireSize, maxWidth, maxHeight int) *Generator {
	return &Generator{
		locationOfChars:   make(map[Point]Char),
		stringOfLocations: make(map[String]Point),
		locationOfStrings: make(map[Point]String),
		width:             0,
		height:            0,
		desireSize:        desireSize,
		maxWidth:          maxWidth,
		maxHeight:         maxHeight,
	}
}

func (g *Generator) Gen(words []string) {
	words = qsort(words)
	//fmt.Println(words)
	firstWord := words[len(words)-1]
	words = words[:len(words)-1]
	s := String{
		root: Point{0, 0},
		s:    firstWord,
		d:    Horizontal,
	}
	g.SetWord(s)
	words = remove(words, firstWord)
	tmp := make([]string, 0)
	lastTmpLength := len(words)

	for {
		if len(words) <= 0 {
			words = make([]string, len(tmp))
			copy(words, tmp)
			if lastTmpLength <= len(words) || len(words) == 0 {
				break
			}
			tmp = tmp[:0]
			lastTmpLength = len(words)
		}
		next := words[len(words)-1]
		l := g.getIntersections(next)
		words = remove(words, next)
		if l == nil || len(l) == 0 {
			tmp = append(tmp, next)
			continue
		}
		b := g.getBestIntersection(l)
		g.SetWord(String{
			root: b.root,
			s:    next,
			d:    b.dir,
		})
	}
}

func (g *Generator) getIntersections(s string) []Intersection {
	list := make([]Intersection, 0)
	for i := 0; i < g.width; i++ {
		for j := 0; j < g.height; j++ {
			p := Point{x: i, y: j}
			c, ok := g.locationOfChars[p]
			if !ok {
				continue
			}
			d := Horizontal + Vertical - c.d

			for x := 0; x < len(s); x++ {
				if string(s[x]) == c.c {
					if ok, l, r := g.checkValidation(s, x, p, d); ok {
						list = append(list, Intersection{
							root:   r,
							points: l,
							dir:    d,
							length: len(s),
						})
					}
				}
			}
		}
	}
	return list
}

func (g *Generator) checkValidation(s string, x int, p Point, d Dir) (bool, []Point, Point) {
	switch d {
	case Horizontal:
		if p.x < x {
			return false, nil, Point{}
		}

		if p.x-x+len(s) > g.maxWidth {
			return false, nil, Point{}
		}
		root := Point{
			x: p.x - x,
			y: p.y,
		}
		points := make([]Point, 0)
		if _, ok := g.locationOfChars[Point{x: root.x - 1, y: root.y}]; ok {
			return false, nil, Point{}
		}
		if _, ok := g.locationOfChars[Point{x: root.x + len(s), y: root.y}]; ok {
			return false, nil, Point{}
		}
		for i := 0; i < len(s); i++ {
			np := Point{
				x: root.x + i,
				y: root.y,
			}
			if c, ok := g.locationOfChars[np]; ok {
				if c.c != string(s[i]) || c.intersection || c.d == d {
					return false, nil, Point{}
				}
				points = append(points, np)
			} else {
				if _, ok1 := g.locationOfChars[Point{x: np.x, y: np.y + 1}]; ok1 {
					if _, ok2 := g.locationOfChars[Point{x: np.x, y: np.y - 1}]; ok2 {
						return false, nil, Point{}
					}
				}
				if _, ok1 := g.locationOfChars[Point{x: np.x, y: np.y + 1}]; ok1 {
					if _, ok2 := g.locationOfChars[Point{x: np.x, y: np.y + 2}]; ok2 {
						return false, nil, Point{}
					}
				}
				if _, ok1 := g.locationOfChars[Point{x: np.x, y: np.y - 1}]; ok1 {
					if _, ok2 := g.locationOfChars[Point{x: np.x, y: np.y - 2}]; ok2 {
						return false, nil, Point{}
					}
				}
			}
		}
		return true, points, root

	case Vertical:
		if p.y < x {
			return false, nil, Point{}
		}
		if p.y-x+len(s) > g.maxHeight {
			return false, nil, Point{}
		}
		root := Point{
			x: p.x,
			y: p.y - x,
		}
		points := make([]Point, 0)
		if _, ok := g.locationOfChars[Point{x: root.x, y: root.y - 1}]; ok {
			return false, nil, Point{}
		}
		if _, ok := g.locationOfChars[Point{x: root.x, y: root.y + len(s)}]; ok {
			return false, nil, Point{}
		}
		for i := 0; i < len(s); i++ {
			np := Point{
				x: root.x,
				y: root.y + i,
			}

			if c, ok := g.locationOfChars[np]; ok {
				if c.c != string(s[i]) || c.intersection || c.d == d {
					return false, nil, Point{}
				}
				points = append(points, np)
			} else {
				if _, ok1 := g.locationOfChars[Point{x: np.x + 1, y: np.y}]; ok1 {
					if _, ok2 := g.locationOfChars[Point{x: np.x - 1, y: np.y}]; ok2 {
						return false, nil, Point{}
					}
				}
				if _, ok1 := g.locationOfChars[Point{x: np.x + 1, y: np.y}]; ok1 {
					if _, ok2 := g.locationOfChars[Point{x: np.x + 2, y: np.y}]; ok2 {
						return false, nil, Point{}
					}
				}
				if _, ok1 := g.locationOfChars[Point{x: np.x - 1, y: np.y}]; ok1 {
					if _, ok2 := g.locationOfChars[Point{x: np.x - 2, y: np.y}]; ok2 {
						return false, nil, Point{}
					}
				}
			}
		}
		return true, points, root
	default:
		panic("direction type doesn't exist!")
	}
}

func (g *Generator) Draw() {
	fmt.Print("   ")
	for j := 0; j < g.width; j++ {
		if j < 10 {
			fmt.Print(" ")
			fmt.Print(j)
			fmt.Print("  ")
		} else {
			fmt.Print(j)
			fmt.Print("  ")
		}
	}
	fmt.Print("\n")

	line := "  +"
	for j := 0; j < g.width; j++ {
		line += "---+"
	}
	line += "\n"
	fmt.Print(line)

	for i := 0; i < g.height; i++ {
		fmt.Print(i)
		if i < 10 {
			fmt.Print(" ")
		}
		fmt.Print("|")
		for j := 0; j < g.width; j++ {
			c, ok := g.locationOfChars[Point{x: j, y: i}]
			if ok {
				fmt.Print(" ")
				fmt.Print(c.c)
				if c.intersection {
					fmt.Print("+")
				} else if c.d == Horizontal {
					fmt.Print(">")
				} else {
					fmt.Print("^")
				}
			} else {
				fmt.Print("   ")
			}
			fmt.Print("|")
		}
		fmt.Print("\n")
		fmt.Print(line)
	}
}

func (g *Generator) getBestIntersection(intersections []Intersection) Intersection {
	bestIntersection := intersections[0]
	for i := 1; i < len(intersections); i++ {
		if g.getScoreOfIntersection(intersections[i]) > g.getScoreOfIntersection(bestIntersection) {
			bestIntersection = intersections[i]
		}
	}
	return bestIntersection
}

func (g *Generator) getScoreOfIntersection(intersection Intersection) int {
	score := len(intersection.points)
	if intersection.dir == Horizontal {
		if g.width < intersection.length+intersection.root.x {
			score = score / (intersection.length + intersection.root.x)
		} else {
			score = score / g.width
		}
		score = score / g.height
	} else if intersection.dir == Vertical {
		if g.height < intersection.length+intersection.root.y {
			score = score / (intersection.length + intersection.root.y)
		} else {
			score = score / g.height
		}
		score = score / g.width
	}

	return score
}

func (g *Generator) SetWord(s String) {
	g.locationOfStrings[s.root] = s
	g.stringOfLocations[s] = s.root
	switch s.d {
	case Horizontal:
		g.width = max(g.width, len(s.s)+s.root.x)
		g.height = max(g.height, 1)
		for i := 0; i < len(s.s); i++ {
			p := Point{x: s.root.x + i, y: s.root.y}
			_, ok := g.locationOfChars[p]
			g.locationOfChars[p] = Char{
				root:         s.root,
				p:            p,
				c:            string(s.s[i]),
				d:            Horizontal,
				master:       s.s,
				intersection: ok,
			}
		}
	case Vertical:
		g.height = max(g.height, len(s.s)+s.root.y)
		g.width = max(g.width, 1)
		for i := 0; i < len(s.s); i++ {
			p := Point{x: s.root.x, y: s.root.y + i}
			_, ok := g.locationOfChars[p]
			g.locationOfChars[p] = Char{
				root:         s.root,
				p:            p,
				c:            string(s.s[i]),
				d:            Vertical,
				master:       s.s,
				intersection: ok,
			}
		}

	default:
		panic("direction type doesn't exist!")

	}
}

func (g *Generator) GetScore() int {
	if g.width*g.height == 0 {
		return -1000000000
	}
	if len(g.stringOfLocations) < g.desireSize {
		return -1000000 + len(g.stringOfLocations)
	}
	score := 0
	checker := make([][]bool, g.height)
	for i := 0; i < g.height; i++ {
		checker[i] = make([]bool, g.width)
		for j := 0; j < g.width; j++ {
			if _, ok := g.locationOfChars[Point{x: j, y: i}]; ok {
				checker[i][j] = true
			}
		}
	}
	for i := 0; i < g.height; i++ {
		for j := 0; j < g.width; j++ {
			if !checker[i][j] {
				counter := g.getNeighbors(Point{x: j, y: i}, checker)
				if counter > 3 {
					score += (counter - 3) * (counter - 3)
				}
			}
		}
	}

	horizontalCount := 0
	verticalCount := 0
	for key, _ := range g.stringOfLocations {
		if key.d == Horizontal {
			horizontalCount += 1
		}
		if key.d == Vertical {
			verticalCount += 1
		}
	}

	intersectionsCount := 0
	for _, value := range g.locationOfChars {
		if value.intersection {
			intersectionsCount += 1
		}
	}
	return - score - int(math.Pow(math.Abs(float64(horizontalCount-verticalCount)), 4)) + int(math.Pow(float64(intersectionsCount), 2))
}

func (g *Generator) getNeighbors(p Point, checker [][]bool) int {
	if p.x < 0 || p.x >= g.width || p.y < 0 || p.y >= g.height {
		return 0
	}

	if checker[p.y][p.x] {
		return 0
	}

	if _, ok := g.locationOfChars[p]; ok {
		return 0
	}

	checker[p.y][p.x] = true
	return 1 + g.getNeighbors(Point{x: p.x + 1, y: p.y,}, checker) +
		g.getNeighbors(Point{x: p.x - 1, y: p.y,}, checker) +
		g.getNeighbors(Point{x: p.x, y: p.y + 1,}, checker) +
		g.getNeighbors(Point{x: p.x, y: p.y - 1,}, checker)
}

func (g *Generator) getNearChars(p Point, d Dir) int {
	return 0
}

func qsort(a []string) []string {
	sort.Slice(a, func(i, j int) bool {
		return len(a[i]) < len(a[j])
	})
	return a
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func remove(arr []string, s string) []string {
	for i := 0; i < len(arr); i++ {
		if arr[i] == s {
			return append(arr[:i], arr[i+1:]...)
		}
	}
	return arr
}

func abs(a int) int {
	if a >= 0 {
		return a
	}
	return -a
}

func randomSubArray(arr []string, l, m int) []string {
	if l > len(arr) {
		panic("????")
	}
	copied := make([]string, len(arr))
	copy(copied, arr)
	for _, a := range arr {
		if len(a) > m {
			copied = remove(copied, a)
		}
	}
	out := make([]string, 0)
	for i := 0; i < l; i++ {
		r := rand.Int() % len(copied)
		out = append(out, copied[r])
		copied = remove(copied, copied[r])
	}
	return out
}

func Start(data string) {
	rand.Seed(time.Now().UnixNano())

	words := make([]string, 0)
	fmt.Println("Reading...")
	file, err := os.Open(data)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		words = append(words, strings.ReplaceAll(scanner.Text(), " ", ""))
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(words))

	desireSize := 5
	maxWidth := 8
	maxHeight := 8
	wordsSize := 7
	interactions := 100

	fmt.Println("Generating...")
	result := NewGenerator(desireSize, maxWidth, maxHeight)
	for i := 0; i < interactions; i++ {
		g := NewGenerator(desireSize, maxWidth, maxHeight)
		w := randomSubArray(words, wordsSize, int(math.Max(float64(maxHeight), float64(maxWidth))))
		g.Gen(w)
		if g.GetScore() > result.GetScore() {
			result = g
		}
	}
	result.Draw()
	fmt.Println(result.GetScore())
}
