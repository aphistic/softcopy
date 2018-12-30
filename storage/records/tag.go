package records

type TagIterator interface {
	Tags() <-chan *TagItem
	Close() error
}

type TagItem struct {
	Tag   *Tag
	Error error
}

type Tag struct {
	ID       int
	Name     string
	Category *TagCategory
	System   bool
}

type TagCategory struct {
	ID   int
	Name string
}
