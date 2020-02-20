package wordsearch

type Generator struct {
	width  int
	height int
	chars  [][]string
}

func NewGenerator(width, height int) *Generator {
	g := &Generator{
		width:  width,
		height: height,
		chars:  make([][]string, height),
	}
	for i := 0; i < height; i++ {
		g.chars[i] = make([]string, width)
	}
	return g
}

func (g *Generator) Gen(words []string) {

}
