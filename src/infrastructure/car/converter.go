package car

import (
	"RentalManagement/logic/model"
	carTypes "github.com/ccsapp/cargotypes"
)

func MapToCarBase(car *carTypes.Car) *model.Car {
	return &model.Car{
		Brand:                  car.Brand,
		DynamicData:            nil,
		Model:                  car.Model,
		TechnicalSpecification: nil,
		Vin:                    car.Vin,
	}
}

func MapToCarAvailable(car *carTypes.Car) *model.CarAvailable {
	return &model.CarAvailable{
		Brand:         car.Brand,
		Model:         car.Model,
		NumberOfSeats: car.TechnicalSpecification.NumberOfSeats,
		Vin:           car.Vin,
	}
}

func MapToCarStatic(car *carTypes.Car) *model.Car {
	carStatic := MapToCarBase(car)
	carStatic.TechnicalSpecification = mapTechnicalSpecification(&car.TechnicalSpecification)
	return carStatic
}

func MapToCar(car *carTypes.Car) *model.Car {
	modelCar := MapToCarStatic(car)
	modelCar.DynamicData = mapDynamicData(&car.DynamicData)
	return modelCar
}

func mapTechnicalSpecification(specification *carTypes.TechnicalSpecification) *model.TechnicalSpecification {
	return &model.TechnicalSpecification{
		Color:         specification.Color,
		Consumption:   specification.Consumption,
		Emissions:     specification.Emissions,
		Engine:        specification.Engine,
		Fuel:          model.TechnicalSpecificationFuel(specification.Fuel),
		FuelCapacity:  specification.FuelCapacity,
		NumberOfDoors: specification.NumberOfDoors,
		NumberOfSeats: specification.NumberOfSeats,
		Transmission:  model.TechnicalSpecificationTransmission(specification.Transmission),
		TrunkVolume:   specification.TrunkVolume,
		Weight:        specification.Weight,
	}
}

func mapDynamicData(data *carTypes.DynamicData) *model.DynamicData {
	return &model.DynamicData{
		DoorsLockState:      model.LockState(data.DoorsLockState),
		EngineState:         model.DynamicDataEngineState(data.EngineState),
		FuelLevelPercentage: data.FuelLevelPercentage,
		Position:            data.Position,
		TrunkLockState:      model.LockState(data.TrunkLockState),
	}
}
