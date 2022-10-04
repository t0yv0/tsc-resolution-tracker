package main

type Node struct {
	label string
	edges []*Node
}

func (n *Node) hasEdge(to *Node) bool {
	for _, e := range n.edges {
		if e == to {
			return true
		}
	}
	return false
}

type Graph struct {
	nodes map[string]*Node
}

func (g *Graph) node(label string) *Node {
	if g.nodes == nil {
		g.nodes = map[string]*Node{}
	}
	n, got := g.nodes[label]
	if !got {
		n = &Node{label: label}
		g.nodes[label] = n
	}
	return n
}

func (g *Graph) addEdge(from, to string) {
	f := g.node(from)
	t := g.node(to)
	f.edges = append(f.edges, t)
}

func (g *Graph) findPath(from, to *Node) *Path {
	return g.findPathWithout(from, to, []*Node{})
}

func (*Graph) contains(nodes []*Node, node *Node) bool {
	for _, n := range nodes {
		if n == node {
			return true
		}
	}
	return false
}

func (g *Graph) findPathWithout(from, to *Node, without []*Node) *Path {
	if g.contains(without, from) {
		return nil
	}
	if g.contains(without, to) {
		return nil
	}
	if from == to {
		return &Path{here: from}
	}
	if from.hasEdge(to) {
		return &Path{here: from, next: &Path{here: to}}
	}
	for _, next := range from.edges {
		if g.contains(without, next) {
			continue
		}
		if p := g.findPathWithout(next, to, append(without, from)); p != nil {
			return &Path{here: from, next: p}
		}
	}
	return nil
}

type Path struct {
	here *Node
	next *Path
}

func (p *Path) slice() []string {
	s := []string{}
	cur := p
	for {
		if cur == nil {
			return s
		}
		s = append(s, cur.here.label)
		cur = cur.next
	}
}
