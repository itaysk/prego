package example2

hello := "universe"

default myrule = false
myrule {
  input.foo == "baz"
}