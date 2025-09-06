# Grammar

The grammar for this implementation is the following

```ebnf
Regex        ::= Alt

Alt          ::= Concat ( "|" Concat )*

Concat       ::= Repeat+

Repeat       ::= Atom Quantifier?

Quantifier   ::= "*" | "+" | "?"

Atom         ::= Literal
               | Wildcard
               | Group

Literal      ::= [a-zA-Z0-9]   (* or define as any non-special char *)

Wildcard     ::= "."

Group        ::= "(" Alt ")"
```