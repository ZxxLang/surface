surface
=======

[API Reference](https://gowalker.org/github.com/ZxxLang/surface)

surface 包主要的代码拷贝自 Go 官方 reflect 包. 对 reflect 包重新组合, 导出更多细节. 这样是为了便于访问属性, 而不是用于更改, 那将造成损失.

**由于使用了极不常规的方法,  有可能造成损失.  慎用 surface.**

缺陷
====

事实上 surface 不能替代官方 reflect 包, surface 中的一些代码依然依赖 reflect 包, surface 没有完全实现 reflect 的功能. 下列函数通过转换可以作为弥补的途径

    func FromValue(v reflect.Value) Value
    Value.ToValue() reflect.Value

很难明确 surface 的应用场景, 通过 surface 增进对 Go 底层数据结构的了解要大于应用意义. surface 尽可能将已知的底层数据结构做了导出处理, 更多的了解这些也许会对您有所启发, 同时您也应该明白在 reflect 包中很多结构是私有的, 这意味着未来官方有可能调整或者算法, 因此适度的使用 surface 是非常有必要的. 

**使用时不要让这些导出无限制的扩散,  那样在未来不容易跟进 官方的变更.**

用例
====

```go
package main

import (
    "fmt"
    "github.com/ZxxLang/surface"
    "reflect"
)

func main() {
    u := User{"Surface", 1}
    rv := reflect.ValueOf(u)
    sv := surface.ValueOf(u)
    fmt.Println(rv.Kind().String(),
        sv.Kind().String())

    fmt.Println(rv.FieldByName("Name").String(),
        sv.Struct().FieldByName("Name").String())

    fmt.Println(rv.FieldByName("Age").Uint(),
        sv.Struct().FieldByName("Age").Uint8()) // must be call Uint8()

    // reflec.Value 和 surface.Value 的相互转换
    fmt.Println(surface.FromValue(rv).Struct().FieldByName("Name").String(),
        sv.Struct().FieldByName("Name").ToValue().String())
}

type User struct {
    Name string
    Age  uint8
}
```

output:
```
struct struct
Surface Surface
1 1
Surface Surface
```

在 reflect.Value 中采用的是 `all in one` 的方法, surface 对非 builtin 类型的进行了分离, `struct` 不是 builtin 类型, 因此需要通过 `Struct()` 进行显示转换. 对于 builtin 类型 surface 也采用了 `all in one` 的方法, 但是类型和对应的方法是独立的, 不像 reflect 包中那样对整型只分 `Int()` 和 `Uint()`.

在方法命名上, surface 围绕 `Type` 和 `Value` 进行命名, 虽然很多名字是一样的, `Type.Array()` 返回的是 `*ArrayType`, 类型的描述. `Value.Array()` 返回的是 `Array` 值的描述. 当然每一个值都包括 `Value.Type` 字段访问.

特别的, `Interface()` 在 reflect 包中返回的是 `interface{}` 值, surface 保留了这个习惯. 同时 surface 中也有 `InterfaceType` 类型和 `Interface` 值的定义, 鉴于名称相同, surface 使用 `Surface()` 加以区别.

贡献
====

问题和建议请至 [Issues](https://github.com/ZxxLang/surface/issues).

如果 surface 对您产生了启发, 也感谢您共享您的代码, 您可以 [Fork](https://github.com/ZxxLang/surface/fork) 并提交您的 `example_goods.go` 样例.

LICENSE
=======
Copyright 2014 The ZxxLang Authors. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.