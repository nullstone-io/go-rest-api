package rest

type DataAccess[TKey any, T any] interface {
	Create(obj T) (*T, error)
	Read(key TKey) (*T, error)
	Update(key TKey, obj T) (*T, error)
	Drop(key TKey) (bool, error)
}
