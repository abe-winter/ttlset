package ttlset
// todo go 1.18 switch to https://pkg.go.dev/golang.org/x/exp/slices

// index of item in slice or -1
func sIndex(item string, slice []string) int {
  for i, v := range slice {
    if v == item {
      return i
    }
  }
  return -1
}

// util to remove from a slice.
func sRemove(item string, slice []string) []string {
  index := sIndex(item, slice)
  if index == -1 {
    return slice
  }
  if len(slice) == 1 {
    return []string{}
  }
  return append(slice[:index], slice[index + 1:]...)
}
