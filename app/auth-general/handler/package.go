package handler

var (
	generalEndpoint string
)

type InitInput struct {
	GeneralEndpoint string
}

func Init(i InitInput) {
	generalEndpoint = i.GeneralEndpoint
}
