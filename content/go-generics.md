---
title: "Understanding Go Generics"
date: 2026-05-10
tags: ["golang", "programming"]
---

# Go Generics

Generics were one of the most anticipated features in Go 1.18. They allow you to write functions and data structures that work with multiple types while maintaining type safety.

## Example: Reverse Slice

Before generics, you'd need a version for every type:

```go
func ReverseInts(s []int) {
    for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
        s[i], s[j] = s[j], s[i]
    }
}
```

With generics:

```go
func Reverse[T any](s []T) {
    for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
        s[i], s[j] = s[j], s[i]
    }
}
```

It's a powerful way to reduce boilerplate!
