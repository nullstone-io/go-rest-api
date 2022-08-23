package rest

type DataAccess[TKey any, T any] interface {
	ParseKey(val string) (TKey, error)
	Read(key TKey) (*T, error)
	Exists(obj T) (bool, error)
	Create(obj T) (*T, error)
	Ensure(obj T) (*T, error)
	Update(key TKey, obj T) (*T, error)
	Drop(key TKey) (bool, error)
}
