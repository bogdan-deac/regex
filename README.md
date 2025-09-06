# Regex

Regex is a toy regex implementation based on automata theory. The pipleine has the following steps

1. Lexing + parsing -> Implemented via a simple recursive descent parser
2. NFA Conversion -> Turns the AST into an NFA using Thompson's algorithm
3. DFA Conversion -> Converts the NFA into a DFA using a classic subset construction algorithm
4. DFA Minimization -> Uses Hopcroft's algorithm

## Supported features

* Basic character recognition - `abcd`
* or operator - `a|b`
* kleene star - `a*`
* plus operator - `a+`
* maybe operator - `a?`
* grouping - `(a|b)*`
* escaped characters - `\||\*`
* wildcards - `.*`


Note - in this implementation, grouping is non-capturing.
Only suports ASCII characters
