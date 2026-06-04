# Coding Standards

## Comments

Keep comments to a minimum. Follow Linus Torvalds' Linux kernel coding style rules for comments:

- **Good code is self-documenting** - If you need a comment to explain what the code does, the code should be rewritten to be clearer.

- **Comments should explain WHY, not WHAT** - Don't comment what the code does; comment why it does it that way, especially if the reason is non-obvious.

- **Avoid redundant comments** - Comments that just repeat what the code says are worse than useless; they get out of sync and become misleading.

- **Document non-obvious behavior** - Only comment things that aren't immediately obvious from reading the code, such as:
  - Why a particular algorithm or approach was chosen
  - Workarounds for bugs or limitations
  - Performance considerations that aren't obvious
  - Business logic that isn't clear from the code itself

- **Keep comments in sync** - If code changes, update or remove comments. Out-of-date comments are worse than no comments.

- **Prefer clear variable/function names over comments** - Instead of commenting what a variable does, rename it to be self-explanatory.

Example of bad comment:
```go
// Loop through all users
for _, user := range users {
    // Process user
    processUser(user)
}
```

Example of good comment:
```go
// Use batch processing to avoid memory issues with large datasets
for _, user := range users {
    processUser(user)
}
```

## Line Length

Keep lines to 80 columns or fewer when possible. This improves readability,
especially when viewing code side-by-side or on smaller screens.

- **Wrap code and comments** - Break long lines across multiple lines to fit within 80 columns
- **Exceptions are acceptable** - Some lines may exceed 80 columns when:
  - Breaking would significantly harm readability
  - Long URLs or import paths cannot be meaningfully split
  - Function signatures are clearer on multi line, with second and more lines tab-indented
  - String literals contain structured data (JSON, SQL, etc.) that's better left intact

Example of wrapping:
```go
// Acceptable: Long function signature on one line
func processUserData(userID int, userName string, userEmail string,
  userRole string, userPermissions []string) error {
}
```
