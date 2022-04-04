# Indigo Guide

The purpose of the guide is to describe how Indigo's rules and the evaluation engine works. We encourage you to read the Indigo source code and examples as primary material, and consider this document as a companion to guide you through the concepts.


Useful links
[Indigo examples](example_test.go)
[CEL examples](/cel/example_test.go)
[CEL Codelabs](https://codelabs.developers.google.com/codelabs/cel-go#0)

***


[1. Introduction](#1-introduction)

   1. [What is a rule?](#what-is-a-rule)
   1. [Why use rules?](#why-use-rules)
   1. [Expressions and rules in Indigo](#expressions-and-rules–in-indigo)

[2. Expression Evaluation](#2-expression-evaluation)

   1. Compilation and Evaluation
   1. Schemas
   1. Data Types 
   1. Boolean Scalar Expressions
   1. Operators
   1. Creating a Rule
   1. Compilation
   1. Input Data
   1. Evaluation 
   1. Evaluation Results 
   1. Short Circuit Evaluation 

[3. Indigo Rules Engine Types](#3-indigo-rules-engine-types)

   1. The Engine type 
   1. The Rule type 

[4. Lists and Maps](#section-4-brlists-and-maps)

   1. Lists
   1. Maps
   1. Functions 

[5. Macros](#5-macros)

[6. Using Protocol Buffers in Rules](#6-using-protocol-buffers-in-rules)

   1. Rules on protocol buffer types 
   1. Field names in rule expressions
   1. Nested types
   1. Enums 

[7. Timestamps and Durations](#7-timestamps-and-durations)

   1. Schema 
   1. Comparisons
   1. Math
   1. Parts of time 


1. Non-boolean Output
1. Rule Hierarchies
1. Diagnostics
   1. Runtime diagnostics 
   1. Compilation output
   
   

1. Nested lists 

</br>
</br>

***
</br>

# 1. Introduction

## What is a rule?
A rule is an expression that can be evaluated to produce an outcome. The outcome may be true or false, or it may be a number, or a string or any other value. The same can be said of any computer language code, but what makes rules different is that their expression language is "configurable" by end users of the software, allowing them to modify how the software works without re-compiling or re-deploying the software. 

Rules are data fed into the software, and the software processes the rules. 

Consider these two equivalent rules, one written in Go, the other in a simplified "rule expression language":


##### Rule in Go
```go
  func qualify(x int) bool {
    if x > 10 {
      return true
    } else {
      return false
	}
  }
```


If the business logic changes, and instead of x > 10 we want x > 15, the Go program must be changed and recompiled. Of course, we could make the value a parameter, but what if instead of simply comparing x to a value, we also want to compare y to another value? We will have to continue changing the qualify function. 

##### Rule in Expression Language
```go
x > 10
```

In the rule language example, we have an expression entered by the user at run-time as a string, and evaluated by the software. We are assuming that the software knows how to interpret the expression and produce a true/false result. If the logic changes, the user can change the rule at run-time without changing the software. 

## Why use rules?
Rules are not appropriate for all situations. For one thing, rules are much slower to evaluate than compiled code. Rules are also more difficult to debug, and because users are able to enter them into the software, they are less predictable than thoroughly tested, quality software. 

So why use them? The primary reason to use rules is that they allow business logic to be configured outside of the software itself. It removes business logic from the software, and makes the software into a data processing *platform*, where the rules are the data. 

A side benefit of using rules in software is that it changes the way engineers think about writing the application. Instead of focusing on the specific business logic that must be implemented, engineers instead think about how to enable *configurable*, *dynamic* business logic that users control. 


## Expressions and rules in Indigo
The Indigo rules engine goes beyond evaluating individual expressions (such as the "rule expression language" example above) and provides a structure and mechanisms to evaluate *groups* of rules arranged in specific ways for specific purposes. Indigo "outsources" the actual expression evaluation to Google's Common Expression Language, though other languages can be plugged in as well. 


</br>
</br>

***
</br>


# 2. Expression Evaluation

We start by looking at how to evaluate individual expressions. In this section we will use the terms "expression" and "rule" interchangeably. Later in this guide we will consider ways to arrange rules in hierarchies for different purposes. 

All of examples use [Google's Common Expression Language](https://github.com/google/cel-go), CEL for short. 


**Note:** In this guide we will touch on the basics of the language, for complete coverage you should read the [CEL language spec](https://github.com/google/cel-spec/blob/master/doc/langdef.md).




See the ``indigo/cel/example_test.go`` file for the code for the examples.

In the code samples below we will provide the corresponding example code like this:
```go
 // indigo/cel/example_test.go:Example_basic()
 ... 
 sample code
 ...
```

This means that the sample code is taken from the function `Example_basic()` in the `indigo/cel/example_test.go` file. 


## Compilation and Evaluation
Indigo supports two-step processing of rules. The first step, compilation, is performed only once, when the rule is created or changed. The second step, evaluation, is performed whenever the software has new data to check against the rule. 

During compilation the rule syntax, functions and data types are checked to ensure the expression is correct. The string is then transformed into an abstract syntax tree that the evaluator will later use to perform the evaluation. The engine only performs the compilation once, but can then evaluate the rules millions of times after. 

## Input Data
In our simple rule example from the introduction, the 'x' was the input data:

```go
x > 10
```

In order to evaluate the rule, we had to provide the value of x in our "real" world. 

In reality, rules may have many pieces of data, such as in this expression:

```go
x > 10 && y != "blue" 
```
Indigo allows you to pass a list of data elements to evaluation. 


## Schemas
One of the strengths of CEL is that it provides type safe rule expressions, and violations of type safety is detected at compile time. This is very useful, since rule-writers get immediate feedback on the validity of their rules. 

Building on our x and y example above , we have to define that in the *schema* for rules that use ``x`` and ``y``, ``x`` is an integer and ``y`` is a string. 

In Indigo, the indigo.Schema type is used to specify the schema:


```go
// indigo/cel/example_test.go:Example_basic()
schema := indigo.Schema{
	Elements: []indigo.DataElement{
		{Name: "x", Type: indigo.Int{}},
		{Name: "y", Type: indigo.String{}},
	},
}
```

The schema is a list of indigo.DataElements. Each data element has a Name, which is the **variable** name that users will use in rule expressions. 

Every rule needs to have a schema, **but a schema can be used by many rules**. As we will see later when evaluating multiple rules, each rule must use the same or compatible schemas. 

### Data types
The data types supported by Indigo are in the `schema.go` file:

- String
- Integer
- Float 
- Bool
- Duration
- Timestamp
- Lists 
- Maps 
- Proto 

Proto is the only type that can have a "nested" structure, similar to a Go struct. Go structs are not supported. 

### Output Schema 

Just as we specify the data types of the input to the rule, we can also specify the data type we expect as a result of rule evaluation. Unlike input schemas which can have many types, the output schema is a single type, and the default is boolean. 

During compilation CEL will verify that the expression produces the type of output you specified in your output schema. 

You specify the output type by setting the ``ResultType`` field on the rule. 

#### Exercises

Modify the example `indigo/cel/example_test.go:Example_basic()` and run ``go test`` in the ``indigo/cel`` directory. 

1. Change the declared data type of `x` to a boolean
1. Change the ResultType of the rule to a string


## Boolean Scalar Expressions 
The simplest type of rule compares scalar values to each other to produce a true/false outcome. 
Our simple x and y rule is such an expression: 

```go
x > 10 && y != "blue" 
```

In some rule engine languages you specify two parts: "if" and "then", like this:

```go
// pseudo-code for a fake rule engine language
if 
   x > 10 && y != "blue" 
then 
   set output = true 
```

CEL derives the value of the output from the value of the expression, in this case a boolean, so the "if/then" construct is not needed. (CEL does support the if/then functionality with the :? operators which we will cover later in the book.)

## Operators

Supported logical operators: ``&& || !``. Parentheses can be used to group expressions. 

Supported comparison operators: `` < <= >= > == != in ``

Supported math operators: `` + - * / % ``


## Creating a rule 

An Indigo rule is a Go struct, so it can be created like this: 


```go
// indigo/cel/example_test.go:Example_basic()
rule := indigo.Rule{
	Schema:     schema,
	ResultType: indigo.Bool{},
	Expr:       `x > 10 && y != "blue"`,
}
```

This initializes the rule, sets its schema, the expected output (boolean) and the expression to evaluate. Indigo defaults the result type to boolean if you don't specify one. 

There is also a convenience function called indigo.NewRule, which also initializes a map of child rules. We'll get to child rules later. 

## Compilation 

Before compiling the rule, we need to create an instance of indigo.DefaultEngine:


```go
// indigo/cel/example_test.go:Example_basic()
engine := indigo.NewEngine(cel.NewEvaluator())
```

This creates a new engine that will use CEL as its expression evaluation language. See the section on Indigo rules engine types for language evaluators and how the Indigo interface types work. 

With an engine, we can now compile the rule:

```go
// indigo/cel/example_test.go:Example_basic()
err := engine.Compile(&rule)
if err != nil {
	fmt.Println(err)
	return
}
```

If there are any compilation errors, the returned error message will tell you what the error is, and where in your expression the error occurred. 

## Input Data

Now that we have a compiled rule, we can start to evaluate data against it. Let's assume we are serving an API that receives data, and it's our job to check the data against the rule and report the results. 

We prepare the data for evaluation like this: 

```go
// indigo/cel/example_test.go:Example_basic()
data := map[string]interface{}{
	"x": 11,
	"y": "red",
}
```

The data is placed in a map, where the key is the variable name you specified in the schema. If we look back at the schema definition from earlier in this section, you'll see that we defined an "x" and a "y" variable:

```go
// Flashback to the schema definition earlier in this section
schema := indigo.Schema{
	Elements: []indigo.DataElement{
		{Name: "x", Type: indigo.Int{}},
		{Name: "y", Type: indigo.String{}},
	},
}
```

A data map does not need to define values for all variables in a schema, but if a variable is used in a rule, it must be in the data map. 

## Evaluation 

We are finally ready to evaluate the input data (x=11, y=red) against the rule to determine if the rule is true or false:

```go
// indigo/cel/example_test.go:Example_basic()
results, err := engine.Eval(context.Background(), &rule, data)
```

The evaluation returns an indigo.Result struct, and an error. It is not an error if the rule returns false; an error is an unexpected issue, such as incorrect data passed. 

Eval accepts a context.Context, and will stop evaluation if the context's deadline expires or is canceled, returning no results and the context's error. 

## Evaluation Results 

The indigo.Result struct contains the output of the expression evaluation, and additional information to help calling code process the results. For now, we'll focus on the boolean Result.ExpressionPass field. This field will tell you if the output of the expression is true or false, which is what we are interested in our example. 

## Short Circuiting 

Many languages, including Go, implement and/or short-circuiting. 

For example, in this statement, the comparison ``b == "blue"`` will never be executed: 

```go
var a *int 
b := "red"

if a!=nil && b == "blue" {
...
}
```

CEL also implements short circuiting, but even allows for a!=nil to be the second clause of an && comparison. See the [table](https://github.com/google/cel-go/blob/master/README.md#partial-state) for examples. 



#### Exercises

Modify the example `indigo/cel/example_test.go:Example_basic()` and run ``go test`` in the ``indigo/cel`` directory. 

1. Comment out the input value for y
   Notice the error message, what does it mean? 

1. Change the input value of x to 7
   Why did this not give an error message? (Although we got false when we wanted true)


</br>
</br>

***
</br>

# 3. Indigo Rules Engine Types

This section is a deeper dive into how Indigo is organized. Knowing this will make it easier to understand how to use Indigo in a larger application.

## The Engine Types

Indigo's rules engine is specified by an interface called Engine, which is composed of two interfaces, a Compiler and an Evaluator. The Compiler interface specifies the Compile method, and Evaluator specifies the Eval method. 

The DefaultEngine struct type implements the Engine interface, and most users will use this type. Since it implements Engine, DefaultEngine implements both Compile and Evaluate. Alternate implementations are possible, including rule evaluators that do not support compilation, for example. 

The engine types are concerned with processing groups of rules organized in a hierarchies. The engine types **do not** concern themselves with the exact nature of the expressions being compiled or evaluated. Instead, they rely on the expression types for that. 

## The Expression Types 
There are two expression interfaces, ExpressionEvaluator, which specifies the evaluation of a single expression, and ExpressionCompiler, which specifies the compilation of a single expression. 

These two interfaces are combined in the ExpressionCompilerEvaluator. 

The indigo/cel.Evaluator implements the ExpressionCompilerEvaluator.

Users of Indigo do not need to interact directly with expression compilation or evaluation; the DefaultEngine handles this. 

## Rule ownership 

Rules are owned by the Go code that calls Indigo engine methods. The indigo.DefaultEngine does **not** store or maintain rules. It is the responsibility of the calling code to be goroutine-safe and to persist rule data for as long as it is needed in the life of the application. 

During compilation, an Engine may update a rule by setting the Program field to the compilation output of the rule. The Engine may require that data later during the evaluation phase. It is the responsibility of the calling code to ensure that the Program data is not tampered with, and that if the rule expression is changed, the rule must be recompiled. 

## Using a Non-CEL Evaluator

There are several Go implementations of scripting languages, such as Javscript implemented by [Otto](https://github.com/robertkrimen/otto), and Lua. These languages are good choices for rule evaluation as well. 

To illustrate how Indigo's interface types work together, here's how you could implement and use a different evaluator such as Otto:

1. In your own module, create a new struct, ``MyEvaluator``
1. Decide if the language requires compilation; if it does, implement the ExpressionCompiler interface, including converting types from Indigo types to the evaluator's types. 
1. Implement the ExpressionEvaluator interface
1. When instantiating indigo.DefaultEngine, pass ``MyEvaluator``

Now, when you call indigo.DefaultEngine.Compile or .Evaluate, it will use your evaluator with your expression language. 



# 4. Lists and Maps
In [section 2](#2-expression-evaluation), we saw how to use scalar values in input data and rule evaluation. Indigo also supports lists and maps in the expression language. These types function much the same as Go slices and maps. 

## Lists 

Below is an example of how to define a list, pass a list in the input data, and how to use a list in a rule expression:


> All of the examples in this section are from the indigo/cel/example_test.go:Example_list()


```go

schema := indigo.Schema{
	Elements: []indigo.DataElement{
		{Name: "grades", Type: indigo.List{ValueType: indigo.Float{}}},
	},
}

rule := indigo.Rule{
	Schema:     schema,
	ResultType: indigo.Bool{},
	Expr:       `size(grades) > 3`,
}

data := map[string]interface{}{
	"grades": []float64{3.4, 3.6, 3.8, 2.9},
}

```

As you can see in the example, when we declare the list variable "grades" in the schema, we need to specify the type of elements we're storing with the ``ValueType`` field. Any Indigo type can be a ``ValueType``. 

In a rule we can access an individual element: 


```go
rule.Expr = `grades[1] == 3.6`
```

With CEL [macros](https://github.com/google/cel-spec/blob/master/doc/langdef.md#macros), we can perform operations on the elements in the list. For example, the ``exists`` macro checks if any of the grades in the list are less than 3.0:

```go
rule.Expr = `grades.exists(g, g < 3.0)`
```

In the macro we define an arbitrary variable g, which will represent each element **value** in the list as CEL executes the macro. 

This rule will evaluate to true, since one of the grades is 2.9.

See the [CEL documentation](https://github.com/google/cel-spec/blob/master/doc/langdef.md#macros) for more macros.


## Maps
Maps work like Go maps, where values are indexed by a key, such as a string. 



> All of the examples in this section are from the indigo/cel/example_test.go:Example_map()


```go
schema := indigo.Schema{
	Elements: []indigo.DataElement{
		{Name: "flights", Type: indigo.Map{KeyType: indigo.String{}, ValueType: indigo.String{}}},
	},
}

data := map[string]interface{}{
   "flights": map[string]string{"UA1500": "On Time", "DL232": "Delayed",   "AA1622": "Delayed"},
}

rule := indigo.Rule{
	Schema:     schema,
	ResultType: indigo.Bool{},
	Expr:       `flights.exists(k, flights[k] == "Delayed")`,
}
```


For maps we have to specify the key type as well as the value type. 

In macros that operate on maps, the value (k) is the map **key**, and we can use that to access values in the map. This will return false, since UA 1500 is delayed:

```go
flights["UA1500"] == "On Time" 

```

# 5. Macros
CEL provides a few macros that provide functionality beyond predicate logic such as == or <=. 

We have already seen one macro (`exists`), but here are some of the macros CEL provides:

- ``has`` checks if a field exists 
- ``all`` will be true if all elements meet the predicate 
- ``exists_one`` will be true if only 1 element matches the 
- ``filter`` can be applied to a list and returns a new list with the matching elements

See the [CEL documentation](https://github.com/google/cel-spec/blob/master/doc/langdef.md#macros) for a complete list of macros. 

There is no macro to do aggregate math on lists or maps. It is recommended to do such calculations outside rule evaluation and provide the data in the input. See [this discussion](https://groups.google.com/g/cel-go-discuss/c/1Y_1APJHk0c/m/JSsKRdGeAQAJ) in the CEL group for more information. 


# 6. Using Protocol Buffers in Rules
Until now all of our rules have operated on simple values, either as individual variables (numbers, strings) or as lists of such simple values. Missing from Indigo and CEL is the ability to use Go structs as input data. (An experimental implementation of struct-support exists in Indigo, but is not recommended). 

Instead, CEL provides support for using Protocol Buffer types in rules. This opens more possibilities for organizing data we pass to rules. If our application serves gRPC APIs, and defines its types with protocol buffers, we can seamlessly pass protocol buffers from our Go code to Indigo rules (and the reverse, as we'll see later). 

The design decisions made in CEL are congruent with those made by protocol buffers, and this is not an accident. CEL natively supports ``google.protobuf.Timestamp`` and ``google.protobuf.Duration``. We'll see more about that later. 

See the [CEL documentation](https://github.com/google/cel-spec/blob/master/doc/langdef.md#protocol-buffer-data-conversion) for how protocol buffers are converted into CEL's internal representation. 


To use protocol buffers in rules, follow Google's tutorial [here](https://developers.google.com/protocol-buffers/docs/gotutorial). In the rest of the guide we'll assume you have a working knowledge of protocol buffers. 

After you've defined your protocol buffers and imported the generated package into your Go code, you can declare an Indigo schema with a protocol buffer type.

Indigo includes a sample protocol buffer [student.proto](../testdata/proto/student.proto), used in all the examples. 

```proto
option go_package = "github.com/ezachrisen/indigo/testdata/school;school";

import "google/protobuf/timestamp.proto";
import "google/protobuf/duration.proto";

message Student {
  double id = 1;
  int32 age = 2;
  double gpa = 3;
  status_type status = 4;
  google.protobuf.Timestamp enrollment_date = 7; 
  
  enum status_type {
	ENROLLED = 0;
	PROBATION = 1;
  }
  
  map<string, string> attrs = 6;
  
  repeated double grades = 5;

  message Suspension {
	string cause = 1;
	google.protobuf.Timestamp date = 2;
  }

  repeated Suspension suspensions = 8;
}
```

A few things to note about this protocol buffer definition:

- It has timestamps (an imported type)
- It has durations (an imported type)
- It has enumerations
- It has maps and lists 
- It has an embedded message type 

Generally, you should assume that any valid protocol buffer can be used in an Indigo rule. 

## Rules on protocol buffer types
Rules on protocol buffer work exactly like rules on simple types such as strings and numbers. 

> The code in this section is from the indigo/cel/example_test.go:Example_protoBasic

In some ways, using protocol buffers instead of native Go types is easier, because in the schema declaration we simply provide an instance of the protocol buffer type we want to use in the ``indigo.Proto`` type. We do not have to specify each field inside the protocol buffer, or the type of values in lists and maps. Nor do we need to define any nested structs or imported types. The protocol buffer definition itself takes care of this for us. 

```go
education := indigo.Schema{
	Elements: []indigo.DataElement{
		{Name: "student", Type: indigo.Proto{Message: &school.Student{}}},
	},
}
data := map[string]interface{}{
	"student": &school.Student{
		Age: 21,
	},
}

rule := indigo.Rule{
	Schema: education,
	Expr:   `student.age > 21`,
}

// Output: false
```


When we declare the schema we provide the address of an instance of the protocol buffer type (``&school.Student{}``). This means that the code that declares a schema has to have access to the generated code. (Later we'll see how to dynamically load protocol buffer types). 

## Field names in rule expressions
In the Go code (where we define the data map), the field ``Age`` is capitalized. That's because protocol buffers are rendered as Go structs where all fields are exported. However, **inside** the rule expression, ``age`` is lower case. That is because **inside** rule expressions, the names are the names used in the ``.proto`` file. So, the field ``enrollment_date`` will be ``EnrollmentDate`` in Go, but ``enrollment_date`` in rule expressions. 