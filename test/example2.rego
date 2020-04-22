package example2

hello := "universe"

default myrule = false
myrule {
  input.foo == "baz"
}

return[res] {
  data.prego_state == "B"
  res := hello
}