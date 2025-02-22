package agent

type Preambles struct {
	One   bool
	Two   bool
	Three bool
	Four  bool
}

func (p *Preambles) Init() *Preambles {
	return &Preambles{One: false, Two: false, Three: false, Four: false}
}

func (p *Preambles) Reset() {
	p.One = false
	p.Two = false
	p.Three = false
	p.Four = false
}