package models

type Todo struct {
	ID   string `bson:"_id,omitempty" json:"id,omitempty"`
	Todo string `bson:"todo" json:"todo"`
	Done bool   `bson:"done" json:"done"`
}
