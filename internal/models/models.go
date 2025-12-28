package models

import "errors"

var RecordNotFound = errors.New("models: no matching record found")
var InvalidMileage = errors.New("models: invalid distance")
