# Regex

Regex is a toy regex implementation based on automata theory. The pipleine has the following steps

1. Lexer -> Processes the initial string into tokens, takes care of escaping
2. Parser -> Turns the tokens into an AST. Uses Pratt parsing, which was the cleanest implementation I found for what I wanted
3. NFA Conversion -> Turns the AST into an NFA using Thompson's algorithm
4. DFA Conversion -> Converts the NFA into a DFA using a classic subset construction algorithm
