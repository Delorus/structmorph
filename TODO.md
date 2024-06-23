* [x] мапить из разных пакетов
* [x] возможность указывать поле из которого брать значение (в том числе для struct в struct)
* [ ] работать с вложенными структурами
* [ ] работать мапами, массивами и слайсами
* [ ] возможность указывать функцию конвертер
* [ ] стандартные конвертеры для классических ситуаций
* [ ] дать возможность мапить поля из нескольких структур
* [ ] дать возможность мапить из функций
* [ ] если не найдено поле по имени, то искать функцию геттер с таким же именем
* [ ] если в одном и том же пакете, то позволить мапить приватные поля
* [ ] ругаться если какое-то поле не замаплено в dst структуре (позволит отлавливать проблемы, когда в src были изменения в полях)
* [ ] особая логика для маппинга с указателя или на указатель с учетом nil
* [ ] различные настройки:
* [ ] - имя конвертеров
* [ ] - имя файла для создания
* [ ] - имя пакета для создания
* [ ] - базовая директория - корень проекта
* [ ] добавить кэш чтобы проходить по файлам только один раз
* [ ] сделать поддержку указания структуры по полному пути
* [ ] возможность добавлять алиасы для from структур

* [ ] дать возможность вместо аннотирования структуры описывать спеку в DSL?
```go

var spec = Spec{
To: PersonDTO,
From: Person,
Mapping: []Mapper{
FromTo("FIO", "FIO")
}
}

```

```go
func Converter1(src Person) PersonDTO { }

func Converter2(src PersonDTO) (Person, error) { }

func Converter3(ctx context.Context, src PersonDTO) (Person, error) { }
```


Нужно ли делать обработку особых случаев или оставить на усмотрение компилятора и goimports?
Особые случаи:
* структура находится в пакете internal
* структура находится в файле который не компилируется 
* структура приватная

Что делать если в проекте есть несколько структур с одинаковыми именами?
Что делать если в проекте есть несколько одинаковых пакетов?