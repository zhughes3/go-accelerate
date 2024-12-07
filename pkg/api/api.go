package api

type ListMeta struct {
	Total int32 `json:"total,omitempty"`
	Count int32 `json:"count,omitempty"`
	Page  int32 `json:"page,omitempty"`
	More  bool  `json:"more,omitempty"`
}

type IdentifiableEntities any

type ListResponse[T IdentifiableEntities] struct {
	Meta    ListMeta `json:"meta,omitempty"`
	Records []T      `json:"records,omitempty"`
}

func NewListResponse[T IdentifiableEntities](records []T) ListResponse[T] {
	return ListResponse[T]{
		Records: records,
	}
}
