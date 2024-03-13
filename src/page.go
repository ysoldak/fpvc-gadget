package main

import "time"

const (
	pageBodyOffset = 24
	pageRowHeight  = 10
)

type Item interface {
	Draw(row int)
}

type ActionalbleItem interface {
	Enter()
}

type Page struct {
	Title string
	items []Item

	subpage bool

	redraw  bool
	cursor  int
	offset  int
	clicked bool

	cycler func(iter int)
}

func NewPage(title string) *Page {
	return &Page{
		Title:  title,
		items:  []Item{},
		redraw: true,
	}
}

// Show page and take control of the encoder
func (p *Page) Enter() {

	for i, item := range p.items {
		if _, ok := item.(ActionalbleItem); ok {
			p.cursor = i
			break
		}
	}

	encoder.SetChangeHandler(p.HandleChange, p.cursor)
	encoder.SetClickHandler(p.HandleClick)

	iter := 0
	for {
		if p.clicked {
			p.clicked = false
			if _, ok := p.items[p.cursor].(*ItemBack); ok {
				return
			}
			p.items[p.cursor].(ActionalbleItem).Enter()
			p.redraw = true
			encoder.SetChangeHandler(p.HandleChange, p.cursor)
			encoder.SetClickHandler(p.HandleClick)
		}
		if p.cycler != nil {
			p.cycler(iter)
		}
		if p.redraw {
			p.redraw = false
			p.Draw()
		}
		time.Sleep(10 * time.Millisecond)

		iter++
		iter %= 100_000 // 1000 sec
	}

}

func (p *Page) Draw() {
	// ensure that ItemBack is the last one when page is a subpage
	if p.subpage {
		if len(p.items) == 0 {
			p.items = append(p.items, NewItemBack())
		}
		if _, ok := p.items[len(p.items)-1].(*ItemBack); !ok {
			p.items = append(p.items, NewItemBack())
		}
		// remove ItemBack from the list if it's not the last one
		for i, item := range p.items {
			if _, ok := item.(*ItemBack); ok && i != len(p.items)-1 {
				p.items = append(p.items[:i], p.items[i+1:]...)
			}
		}
	}
	// draw page
	display.Clear()
	// draw title
	display.Print(2, 6, p.Title)
	display.Line(2, 11, 125, 11, WHITE)
	// draw items
	for i, item := range p.items {
		if i < p.offset || i > p.offset+3 {
			continue
		}
		item.Draw(i - p.offset)
	}
	display.Print(0, pageBodyOffset+int16(p.cursor-p.offset)*pageRowHeight, ">")
	display.Show()
}

func (p *Page) HandleClick() {
	if len(p.items) == 0 {
		return
	}
	p.clicked = true
}

func (p *Page) HandleChange(value int) int {
	if len(p.items) == 0 {
		return 0
	}
	// s.cursorPrev = s.cursor
	display.Erase(0, pageBodyOffset+int16(p.cursor-p.offset)*pageRowHeight, ">")

	dir := value - p.cursor
	// p.cursor = value

	for {
		p.cursor += dir
		if p.cursor < 0 {
			p.cursor = len(p.items) - 1
		}
		if p.cursor > len(p.items)-1 {
			p.cursor = 0
		}
		if _, ok := p.items[p.cursor].(ActionalbleItem); ok {
			break
		}
	}

	if p.cursor < p.offset {
		if p.cursor < 3 {
			p.offset = 0
		} else {
			p.offset = p.cursor
		}
		p.redraw = true
	}
	if p.cursor > p.offset+3 {
		p.offset = p.cursor - 3
		p.redraw = true
	}

	display.Print(0, pageBodyOffset+int16(p.cursor-p.offset)*pageRowHeight, ">")
	display.Show()

	return p.cursor
}

// ----------------------------

type ItemSimple struct {
	Title string
}

func NewItemSimple(title string) *ItemSimple {
	return &ItemSimple{
		Title: title,
	}
}

func (is *ItemSimple) Draw(row int) {
	display.Print(2, pageBodyOffset+int16(row)*pageRowHeight, " "+is.Title)
}

// ----------------------------

type ItemBack struct {
	ItemSimple
}

func NewItemBack() *ItemBack {
	return &ItemBack{
		ItemSimple: ItemSimple{
			Title: "Back",
		},
	}
}

func (ib *ItemBack) Enter() {
}
