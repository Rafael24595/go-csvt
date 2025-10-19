# Go CSVT

**go-csvt** is a Go library for serializing and deserializing CSVT, a table-based hierarchical data format inspired by CSV but designed to represent multiple related tables within a single document.

## What is CSVT?

CSVT (Comma Separated Value Tables) extends the traditional CSV concept by allowing multiple named tables with row references. Each table has a header and rows, and values can include references to other tables using the format `$Table&Key_Index`.

### Table Headers

- **Root table**: begins with `/**` followed by the table name and identifier hash.  
- **Secondary tables**: begin with `///` followed by the table name and identifier.  
- Headers within a table start with `H->` followed by `;`-separated column names.

### Row Prefixes

- Each row starts with its index followed by `->` (e.g., `0->`, `1->`).  
- **Row 0** may contain default values such as empty maps, empty arrays, or empty strings.  

### Reserved Table Names

Some secondary tables have predefined roles:

- `common-array` – stores slices (arrays) used by other tables.  
- `common-map` – stores maps (key-value structures) used by other tables.  
- These tables typically use **row 0** to store default or empty values.

### Format Delimiters and Symbols

The CSVT format relies on a series of tokens to structure headers, rows, maps, arrays, strings, and references.  
The following constants are defined in the library:

| Token | Meaning / Function |
|-------|--------------------|
| /** | Denotes the beginning of the root table with its name and hash. |
| /// | Denotes the beginning of a secondary table with its name and identifier. |
| H-> | Indicates a header line prefix within a table. |
| N-> | Prefix for row lines within a table (e.g., 0->, 1->, etc.). Row 0 may contain default values (empty maps, arrays, or strings). |
| ; | Separates fields in headers (e.g., between column names) and structure row values. |
| , | Separates elements inside maps and arrays. |
| = | Links key and value inside a map (e.g., key=value). |
| ^ | Marks the end of a map structure. |
| \| | Marks the end of an array structure. |
| : | Marks the end of a structure. |
| $ | Indicates a reference to another table (e.g., $Table&Key_Index). |
| _ | Separates the reference identifier from its positional index (e.g., Key_Index). |


### Example:

```text
/** Context&8d920b08e3111b654da8e779052b9aeda392d0a4
H-> Id;Status;Timestamp;Dictionary;Owner;Modified
0-> "d884c662-1242-4015-9b2b-425d408c0154";true;1747123293451;$common-map_1;"rafael";1747123293451:

/// common-map
H-> 
0-> ^
1-> "id"=$Item&8d920b08e3111b654da8e779052b9aeda392d0a4_0,"server-0"=$Item&8d920b08e3111b654da8e779052b9aeda392d0a4_1^

/// Item&8d920b08e3111b654da8e779052b9aeda392d0a4
H-> Order;Private;Status;Value
0-> 1;false;true;"67effe32-08d1-480d-ab9d-486c6d3bf637":
1-> 0;false;true;"http://localhost:8080":
```

## Installation

```bash
go get github.com/Rafael24595/go-csvt
```

## Basic Usage

### Parsing a CSVT document:

```go
buffer, err := utils.ReadFile("path/to/file.csvt")
if err != nil {
  return nil, err
}

if len(buffer) == 0 {
  return make(map[string]T), nil
}

var vector []T
err = csvt.Unmarshal(buffer, &vector)
if err != nil {
  return nil, err
}
```

### Serializing a document:

```go
result, err := csvt.Marshal(items...)
if err != nil {
  return err
}

return utils.WriteFile(m.path, result)
```
