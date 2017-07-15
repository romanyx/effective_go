## Управляющие структуры ## {#control-structures}

Управляющие структуры в Go делают тоже что и подобные им структуры в C, но отличаются от них важными особенностями. Циклы `do` и `while` отсутстуют, вместо них исользуется обобщенный цикл `for`; `switch` более гибок; `if` и `switc`h могут принимать необязательное условие до начала выполнения конструкции; конструкции `break` и `continue` могут принимать метку для того чтобы знать что им нужно прекратить или продолжить; добавленны новые управляющие структуры, такие как типизированный `switch` и мультипликсирующий `select`.

### If ### {#if}

В Go структура `if` в простой форме выглядит следующим образом

``` go
if x > 0 {
	return y
}
```

Обязательные фигурные скобки деляют необходиомой запись конструкцию `if` в несколько строк. В любом случае это хороший стиль записи, особенно если тело структуры содержит управляющий оператор такой как `return` или `break`

Конструкции `if` и `switch` принимают необязательное начальное условие, и обычно используются для определения локальной переменной.

``` go
if err := file.Chmod(0664); err != nil {
	log.Print(err)
	return err
}
```

В библиотеках Go часто встречается запись конструкции `if` в которой тело конструкции заканчивается оператором `break`, `continue`, `goto` или `return`, т.е. выполнение последующих конструкций обрывается вызовом оператора, данная форма записи позволяет отказаться от использования `else`.

``` go
f, err := os.Open(name)
if err != nil {
	return err
}
codeUsing(f)
```

Это общий пример ситуации, в которой код защищает от возникающих в дальнейшем ошибок. Код удобно читать когда поток конструкций идет последовательно в один ряд и реагирование на ошибки происходит как только они возникают. Так как появление ошибки приводит к прерыванию исполнения последующего кода вызовом оператора `return`, то конструкция `else` не требуется.

``` go
f, err := os.Open(name)
if err != nil {
    	return err
}
d, err := f.Stat()
if err != nil {
    	f.Close()
	return err
}
codeUsing(f, d)
```

### Redeclaration and reassignment ### {#redeclaration}

An aside: The last example in the previous section demonstrates a detail of how the `:=` short declaration form works. The declaration that calls `os.Open` reads,

``` go
f, err := os.Open(name)
```

This statement declares two variables, `f` and `err`. A few lines later, the call to `f.Stat` reads,

``` go
d, err := f.Stat()
```

which looks as if it declares `d` and `err`. Notice, though, that `err` appears in both statements. This duplication is legal: `err` is declared by the first statement, but only _re-assigned_ in the second. This means that the call to `f.Stat` uses the existing `err` variable declared above, and just gives it a new value.

In a `:=` declaration a variable `v` may appear even if it has already been declared, provided:

*   this declaration is in the same scope as the existing declaration of `v` (if `v` is already declared in an outer scope, the declaration will create a new variable §),
*   the corresponding value in the initialization is assignable to `v`, and
*   there is at least one other variable in the declaration that is being declared anew.

This unusual property is pure pragmatism, making it easy to use a single `err` value, for example, in a long `if-else` chain. You'll see it used often.

§ It's worth noting here that in Go the scope of function parameters and return values is the same as the function body, even though they appear lexically outside the braces that enclose the body.

### For ### {#for}

The Go `for` loop is similar to—but not the same as—C's. It unifies `for` and `while` and there is no `do-while`. There are three forms, only one of which has semicolons.

``` go
// Like a C for
for init; condition; post { }

// Like a C while
for condition { }

// Like a C for(;;)
for { }
```

Short declarations make it easy to declare the index variable right in the loop.

``` go
sum := 0
for i := 0; i < 10; i++ {
    sum += i
}
```

If you're looping over an array, slice, string, or map, or reading from a channel, a `range` clause can manage the loop.

``` go
for key, value := range oldMap {
    newMap[key] = value
}
```

If you only need the first item in the range (the key or index), drop the second:

``` go
for key := range m {
    if key.expired() {
        delete(m, key)
    }
}
```

If you only need the second item in the range (the value), use the _blank identifier_, an underscore, to discard the first:

``` go
sum := 0
for _, value := range array {
    sum += value
}
```

The blank identifier has many uses, as described in [a later section](#blank).

For strings, the `range` does more work for you, breaking out individual Unicode code points by parsing the UTF-8. Erroneous encodings consume one byte and produce the replacement rune U+FFFD. (The name (with associated builtin type) `rune` is Go terminology for a single Unicode code point. See [the language specification](https://golang.org/ref/spec#Rune_literals) for details.) The loop

``` go
for pos, char := range "日本\x80語" { // \x80 is an illegal UTF-8 encoding
    fmt.Printf("character %#U starts at byte position %d\n", char, pos)
}
```

prints

``` go
character U+65E5 '日' starts at byte position 0
character U+672C '本' starts at byte position 3
character U+FFFD '�' starts at byte position 6
character U+8A9E '語' starts at byte position 7
```

Finally, Go has no comma operator and `++` and `--` are statements not expressions. Thus if you want to run multiple variables in a `for` you should use parallel assignment (although that precludes `++` and `--`).

``` go
// Reverse a
for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
    a[i], a[j] = a[j], a[i]
}
```

### Switch ### {#switch}

Go's `switch` is more general than C's. The expressions need not be constants or even integers, the cases are evaluated top to bottom until a match is found, and if the `switch` has no expression it switches on `true`. It's therefore possible—and idiomatic—to write an `if`-`else`-`if`-`else` chain as a `switch`.

``` go
func unhex(c byte) byte {
    switch {
    case '0' <= c && c <= '9':
        return c - '0'
    case 'a' <= c && c <= 'f':
        return c - 'a' + 10
    case 'A' <= c && c <= 'F':
        return c - 'A' + 10
    }
    return 0
}
```

There is no automatic fall through, but cases can be presented in comma-separated lists.

``` go
func shouldEscape(c byte) bool {
    switch c {
    case ' ', '?', '&', '=', '#', '+', '%':
        return true
    }
    return false
}
```

Although they are not nearly as common in Go as some other C-like languages, `break` statements can be used to terminate a `switch` early. Sometimes, though, it's necessary to break out of a surrounding loop, not the switch, and in Go that can be accomplished by putting a label on the loop and "breaking" to that label. This example shows both uses.

``` go
Loop:
	for n := 0; n < len(src); n += size {
		switch {
		case src[n] < sizeOne:
			if validateOnly {
				break
			}
			size = 1
			update(src[n])

		case src[n] < sizeTwo:
			if n+1 >= len(src) {
				err = errShortInput
				break Loop
			}
			if validateOnly {
				break
			}
			size = 2
			update(src[n] + src[n+1]<<shift)
		}
	}
```

Of course, the `continue` statement also accepts an optional label but it applies only to loops.

To close this section, here's a comparison routine for byte slices that uses two `switch` statements:

``` go
// Compare returns an integer comparing the two byte slices,
// lexicographically.
// The result will be 0 if a == b, -1 if a < b, and +1 if a > b
func Compare(a, b []byte) int {
    for i := 0; i < len(a) && i < len(b); i++ {
        switch {
        case a[i] > b[i]:
            return 1
        case a[i] < b[i]:
            return -1
        }
    }
    switch {
    case len(a) > len(b):
        return 1
    case len(a) < len(b):
        return -1
    }
    return 0
}
```

### Type switch ### {#type_switch}

A switch can also be used to discover the dynamic type of an interface variable. Such a _type switch_ uses the syntax of a type assertion with the keyword `type` inside the parentheses. If the switch declares a variable in the expression, the variable will have the corresponding type in each clause. It's also idiomatic to reuse the name in such cases, in effect declaring a new variable with the same name but a different type in each case.

``` go
var t interface{}
t = functionOfSomeType()
switch t := t.(type) {
default:
    fmt.Printf("unexpected type %T\n", t)     // %T prints whatever type t has
case bool:
    fmt.Printf("boolean %t\n", t)             // t has type bool
case int:
    fmt.Printf("integer %d\n", t)             // t has type int
case *bool:
    fmt.Printf("pointer to boolean %t\n", *t) // t has type *bool
case *int:
    fmt.Printf("pointer to integer %d\n", *t) // t has type *int
}
```
