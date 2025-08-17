package domain

import "errors"

// ErrActionCompletedEarly ошибка: действие было выполнено ранее
var ErrActionCompletedEarly = errors.New("action completed early")
var ErrNoContent = errors.New("action completed early")
