# Monkey

Monkey is a programming language specifically designed for learning how interpreters and compilers
work. It is from Thorsten Ball's series of books, found [here](https://thorstenball.com/books/).
  
Expressed as a list of features, Monkey has the following:  
 - C-like syntax
 - Variable bindings
 - intergers and booleans
 - arithmetic expressions
 - built-in functions
 - first-class and higher-order functions
 - closures
 - a string data structure
 - an array data structure
 - a hash data structure

---

Here is how we bind values to names in Monkey:

```
let age = 1;
let name = "Monkey";
let result = 10 * (20 / 2);
```

Here is what binding an array of integers to a name looks like:

```
let myArray = [1, 2, 3, 4, 5];
```

Here is a hash, wehre values are associated with keys:

```
let ricardo = {"name": "Ricardo", "age": 21};
```

Accessing the elements in arrays and hashes is done with index expressions:

```
myArray[0]      // -> 1 
ricardo["name"] // -> "Ricardo"
```

The `let` statement can also be used to bind functions to names. Here's a small funciton that 
adds two numbers:

```
let add = fn(a, b) { return a + b; };
```

But Monkey not only supports `return` statements. Implicit returns values are also possible, which
means we can leave out the `return` if we want to:

```
let add = fn(a, b) { a + b; };
```

And calling a function is as easy as you'd expect:

```
add(1, 2);
```

A more complex function, such as a `fibonacci` function that returns the nth Fibonacci number, might
look like this:

```
let fibonacci = fn (x) {
    if (x == 0) {
        0
    } else {
        if (x == 1) {
            1
        } else {
            fibonacci(x - 1) + fibonacci(x - 2);
        }
    }
};
```

Note the recursive calls to `fibonacci` itself!  

Here is an example of Monkey's higher order functions:

```
let twice = fn(f, x) {
    return f(f(x));
};

let addTwo = fn(x) {
    return x + 2;
};

twice(addTwo, 2); // -> 6
```
