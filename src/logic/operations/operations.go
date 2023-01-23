package operations

import "RentalManagement/infrastructure/car"

type operations struct {
	carClient car.ClientWithResponsesInterface
}

func NewOperations(carClient car.ClientWithResponsesInterface) IOperations {
	return &operations{
		carClient: carClient,
	}
}
