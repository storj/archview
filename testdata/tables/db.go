package tables

// Postgres is the database implementation; its tables need no Go wrapper types.
// architecture: Database Implementation
type Postgres struct{}

// architecture:link Postgres -> external:users : reads and writes
// architecture:link Postgres -> external:comments : reads and writes
// architecture:link external:comments -> external:users : references
// architecture:resource external:users : Database Table
// architecture:resource external:comments : Database Table
