# PECK

A simple REST API for SQL databases. Its intent is to enable the use of SQL while abstracting away the need for specific database drivers, to simplify common CRUD operations with REST calls, and handle errors and transactions automatically.


Peck was not designed to be used by clients directly. but was designed to be used by an internal API.

```
[Client] <--> [API] <--> [Peck] <--> [database]
```


### Authentication

All of Pecks interfaces require a SQL username and password. The usename and password must be provided in the `Authorization` header using the `Basic RFC 7617` scheme in the http request. Peck does not store credentials. All operations are performed with the username and password provide within a callers request.


### ContentTypes


#### SQL

All SQL commands must be embeded in the HTTP request body and be encoded in UTF-8. The request header must have a `ContentType` of `application/sql`.

#### JSON

All JSON commands must be embeded in the HTTP request body and be encoded in UTF-8. all requests header must have a `ContentType` of `application/json`.

### Invoking SQL Queries

You can invoke SQL queries through this interface. only statements that return rows are permitted. If any changes are made to a tables data through this interface, they will be rolled back. This interface will *always* return a JSON array where each element in the array is a object representing a row. If a query returns no rows, an empty array will be returned.

`POST /<database>/query`

### Invoking SQL Statements

You can invoke SQL statements that modify table data with this interface. this interface will return a JSON object that describes the number rows modified or any errors that occured durring the request.

`POST /<database>/modify`

### Getting a Row

If your table has a primary key or some other unique constraint you can select a single row by specifying it in the URL's Query. Specifying a query should *always* select a single row. A query that selects more then a one row is considered an error. This interface will return either a JSON object that represents a single row or an error. 

`GET /database/schema/table?column=value&column=value`

This request is synonymous with the folowing SQL query.

```sql
-- TSQL
SELECT *
FROM [database].[schema].[table]
WHERE 
    [column] = 'value'
    AND [column] = 'value'
```


### Getting many rows

Peck does not try to replace SQL queries. SQL is a powerfull search language. Peck's focus is to make menial tasks such as, getting, inserting, and modifying rows indevidually simple. If you need to get mulltiple rows then you will probably want to use the SQL query API mentioned above.

That said there are times when you would like to retreive multiple rows with a known index, like a foreign key. it is common to want to retreive *all* rows with a matching foreign key. for this Peck has the Join interface.

`GET /<database>/<schema>/<table>/join?<column>=<value>`

The Join interface always returns a JSON array of objects where each object represents a row in the table with a matching foreign key. this interface will return an empty array if there are no rows with a matching foreign key.

### Inserting and updating

Peck can abstract away inserting or updating data, and simply perform the correct operation on the data you provide it.

`PUT /<database>/<schema>/<table>?<column>=<value>`

```json 
{"column":"value","column":"value"}
```

- To perform a insert either provide a JSON object with a new key/value pair that satisfies a unique constraint.
- To update a row, provide a JSON object with key/values that match an existing unique index in the table.

This interface requires that you identify the resource within the URL query. If a request to this interface does not result in a insert or an update it is considered an error. If an error occurs all changes made during the request will be rolled back. This interface will always return a JSON object that describes either the modification to a row or an error. 


#### insert/update many rows

Peck can manage batch insert/updates, however for a row to be updated, the JSON object must contain a key/value that matches an existing unique index, otherwise it will be inserted as a new row. If a request to this interface does not result in a insert or an update it is considered an error. If an error occurs all changes made during the request will be rolled back. This interface will always return a JSON object that describes either the modifications to rows or an error.

`POST /<database>/<schema>/<table>`

```json
[
    {"column":"value","column":"value"},
    {"column":"value","column":"value"}
]
```

### Deleting Data

#### delete a single row

You can delete a single row by specifying its unique index within the URL query. If this interface does not delete one row it is considered an error. If an error occurs all changes made during the request will be rolled back. This interface will always return a JSON object that describes either the deletion of a row or an error.

`DELETE /<database>/<schema>/<table>?<column>=<value>`



