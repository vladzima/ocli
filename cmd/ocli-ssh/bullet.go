package main

import (
	"github.com/google/uuid"
)

type BulletColor int

const (
	ColorDefault BulletColor = iota
	ColorBlue
	ColorGreen
	ColorYellow
	ColorRed
)

type Bullet struct {
	ID        string
	Content   string
	Children  []*Bullet
	Parent    *Bullet
	Collapsed bool
	IsEditing bool
	Color     BulletColor
	IsTask    bool
	Completed bool
}

func NewBullet(content string) *Bullet {
	return &Bullet{
		ID:        uuid.New().String(),
		Content:   content,
		Children:  make([]*Bullet, 0),
		Collapsed: false,
		IsEditing: false,
		Color:     ColorDefault,
		IsTask:    false,
		Completed: false,
	}
}

func (b *Bullet) AddChild(child *Bullet) {
	child.Parent = b
	b.Children = append(b.Children, child)
}

func (b *Bullet) RemoveChild(child *Bullet) {
	for i, c := range b.Children {
		if c.ID == child.ID {
			b.Children = append(b.Children[:i], b.Children[i+1:]...)
			child.Parent = nil
			break
		}
	}
}

func (b *Bullet) InsertChildAt(index int, child *Bullet) {
	if index < 0 || index > len(b.Children) {
		return
	}
	child.Parent = b
	b.Children = append(b.Children[:index], append([]*Bullet{child}, b.Children[index:]...)...)
}

func (b *Bullet) GetDepth() int {
	depth := 0
	current := b
	for current.Parent != nil {
		depth++
		current = current.Parent
	}
	return depth
}

func (b *Bullet) GetDepthFrom(ancestor *Bullet) int {
	if b == ancestor {
		return 0
	}
	depth := 0
	current := b
	for current.Parent != nil {
		if current.Parent == ancestor {
			return depth + 1
		}
		depth++
		current = current.Parent
	}
	return depth
}

func (b *Bullet) IsVisible() bool {
	parent := b.Parent
	for parent != nil {
		if parent.Collapsed {
			return false
		}
		parent = parent.Parent
	}
	return true
}

func (b *Bullet) Toggle() {
	if len(b.Children) > 0 {
		b.Collapsed = !b.Collapsed
	}
}

func (b *Bullet) GetVisibleDescendants() []*Bullet {
	var visible []*Bullet
	if !b.Collapsed {
		for _, child := range b.Children {
			visible = append(visible, child)
			visible = append(visible, child.GetVisibleDescendants()...)
		}
	}
	return visible
}

func (b *Bullet) CycleColor() {
	b.Color = (b.Color + 1) % 5
}

func (b *Bullet) ToggleTask() {
	b.IsTask = !b.IsTask
	if !b.IsTask {
		b.Completed = false
	}
}

func (b *Bullet) ToggleComplete() {
	if b.IsTask {
		b.Completed = !b.Completed
	}
}
