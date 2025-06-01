# zerologgorm

`zerologgorm` is a GORM logger implementation that uses the [zerolog](https://github.com/rs/zerolog) library for fast and flexible JSON logging. This library allows you to seamlessly integrate GORM's logging capabilities with zerolog, providing a structured and efficient way to log your database interactions.

## Features

*   **Fast JSON Logging:** Leverages zerolog for high-performance structured logging.
*   **GORM Integration:** Provides a `logger.Interface` implementation compatible with GORM.
*   **Configurable:** Offers various options to customize logging behavior:
    *   Log SQL parameters.
    *   Ignore `ErrRecordNotFound` errors.
    *   Skip caller frames for cleaner logs.
    *   Customize the SQL field name.
    *   Set default log levels.
    *   Define a slow query threshold.
*   **Context-Aware Logging:** Propagates context information to zerolog.

## Installation

To use `zerologgorm` in your Go project, you need to have Go installed (version 1.16 or higher is recommended).

You can install the library using `go get`:

```bash
go get github.com/your-repo-path/zerologgorm
```
*(Adjust `github.com/your-repo-path/zerologgorm` to the actual import path of your library.)*

## Usage

Here's how you can initialize and use `zerologgorm` with GORM:

```go
package main

import (
	"context" // Added import for context
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/your-repo-path/zerologgorm" // Adjust import path
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Create a zerolog logger
	zerologLogger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Initialize GORM with zerologgorm
	// Initialize with options:
	newLogger := zerologgorm.NewLogger(
		zerologgorm.WithDefaultLogLevel(zerolog.DebugLevel), // Set default log level for GORM traces
		zerologgorm.WithSlowThreshold(200*time.Millisecond),  // Log queries slower than 200ms as warnings
		zerologgorm.WithLogParams(),                         // Log SQL parameters
		zerologgorm.WithIgnoreNotFoundError(),               // Ignore ErrRecordNotFound
		// zerologgorm.WithSkipFrames(1),                    // Optional: Skip caller frames
		// zerologgorm.WithSqlFieldName("custom_sql_field"), // Optional: Custom SQL field name
	)

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		// Log the error using your zerolog instance if needed
		zerologLogger.Fatal().Err(err).Msg("failed to connect database")
	}

	// Your GORM models and operations
	type Product struct {
		gorm.Model
		Code  string
		Price uint
	}

	// Migrate the schema
	// Use a context with your zerolog logger for context-aware logging
	ctx := zerologLogger.WithContext(context.Background())
	db.WithContext(ctx).AutoMigrate(&Product{})

	// Create
	db.WithContext(ctx).Create(&Product{Code: "D42", Price: 100})

	// Read
	var product Product
	db.WithContext(ctx).First(&product, 1) // find product with integer primary key
	db.WithContext(ctx).First(&product, "code = ?", "D42") // find product with code D42

	// Update - update product's price to 200
	db.WithContext(ctx).Model(&product).Update("Price", 200)

	// Delete - delete product
	db.WithContext(ctx).Delete(&product, 1)
}
```
*Note: The example above uses `context.Background()`. In a real application, you should use the appropriate context for your requests or operations.*

## Configuration Options

`zerologgorm` provides several options to customize the logger's behavior. These options are passed to the `NewLogger` function:

*   `WithDefaultLogLevel(level zerolog.Level)`: Sets the default log level for GORM trace messages (e.g., SQL queries). The default is `zerolog.DebugLevel`.
    *   Example: `zerologgorm.WithDefaultLogLevel(zerolog.InfoLevel)`
*   `WithSlowThreshold(threshold time.Duration)`: Sets the duration after which a query is considered "slow". Slow queries are logged at `zerolog.WarnLevel`. The default is `500 * time.Millisecond`.
    *   Example: `zerologgorm.WithSlowThreshold(250 * time.Millisecond)`
*   `WithLogParams()`: Enables logging of SQL query parameters. By default, parameters are not logged for security and verbosity reasons.
    *   Example: `zerologgorm.WithLogParams()`
*   `WithIgnoreNotFoundError()`: Prevents GORM's `ErrRecordNotFound` error from being logged during trace operations. Other errors will still be logged.
    *   Example: `zerologgorm.WithIgnoreNotFoundError()`
*   `WithSkipFrames(skip int)`: Configures the logger to skip the specified number of caller frames when reporting the source file and line number. This is useful if you have wrapper functions around GORM calls and want the log to point to the original caller.
    *   Example: `zerologgorm.WithSkipFrames(1)`
*   `WithSqlFieldName(name string)`: Sets the field name used for logging the SQL query string in trace messages. The default is `"sql"`.
    *   Example: `zerologgorm.WithSqlFieldName("query_string")`

## Contributing

Contributions are welcome! If you find a bug or have a feature request, please open an issue on the GitHub repository.

If you'd like to contribute code:

1.  Fork the repository.
2.  Create a new branch for your feature or bug fix (`git checkout -b feature/your-feature-name` or `git checkout -b fix/your-bug-fix`).
3.  Make your changes and commit them with clear and concise messages.
4.  Ensure your code is formatted with `go fmt` and passes `go vet` and `go test`.
5.  Push your changes to your fork (`git push origin feature/your-feature-name`).
6.  Open a pull request to the main repository.

Please ensure your code follows the existing style.

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.
*(Ensure you have a LICENSE file in your repository, typically containing the MIT License text.)*
