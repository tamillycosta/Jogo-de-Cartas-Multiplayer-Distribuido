package repository

//interface genérica para CRUD básico
type Repository[T any] interface{
	Create(data *T) (*T, error)
	CreateWithID(id string, data *T) (*T, error)
	GetAll() ([]*T, error)
	DeleteAll() error
	Delete(id string)
	Update(data *T) (*T, error)
}
