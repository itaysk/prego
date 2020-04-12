package example

hello := "world"

default myrule = false
myrule {
  input.foo == "bar"
}