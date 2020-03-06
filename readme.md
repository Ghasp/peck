# PECK

A simple REST API for SQL databases. Its intent is to enable the use of SQL while abstracting away the need for specific database drivers, to simplify common CRUD operations with REST calls, and handle errors and transactions automatically.

Peck was not designed to be used by clients directly.

```
[Client] <--> [API] <--> [Peck] <--> [database]
```


## Dependancies

- pyodbc
- aiohttp

```bash
pip install pyodbc aiohttp
```


## Start

```bash
python peck.py
```

### Authentication

All of Pecks interfaces require a sql username and password. The usename and password must be provided in the `Authorization` header using the `Basic RFC 7617` scheme in the http request. Peck can not, on its own, authenticate to a database. All operations are performed with the username and password provide within a callers request.


### Invoking SQL Commands

All sql commands must be embeded in the HTTP request body and be encoded in UTF-8. The request header must have a `ContentType` of `application/sql`.

#### Query 

You can invoke sql queries through this interface. only statements that return rows are permitted. If any changes are made to a tables data through this interface, they will be rolled back. This interface will always return a JSON array where each element in the array is a object representing a row. If a query returns no rows, an empty array will be returned.

`POST /{database}/query`

#### Execute

You can invoke sql statements that modify table data with this interface. this interface will return a JSON object that describes the number rows modified or any errors that occured durring the request.

`POST /{database}/execute`

### Getting a Row

If your table has a primary key or some other unique constraint you can select a single row by specifying it in the URL's fragment. Specifying a fragment should *always* select a single row. A fragment that selects more then a one row is considered an error. This interface will return either a JSON object that represents a single row or an error. 

`GET /database/schema/table#column=value&column=value`

This request is synonymous with the folowing sql query.

```sql
-- TSQL
SELECT *
FROM [database].[schema].[table]
WHERE 
    [column] = 'value'
    AND [column] = 'value'
```


### Getting Many Rows

SQL is a powerful search language, Peck's search interface is not inteded to replace it. However many common SQL operations are simple enough that they can be represented in a URL's query. Query parameters such as `order`, `limit`, and `offset` allow you to paginate through a tables rows. The `filter` parameter can be used to filter the rows of a table. 

When a URL fragment is not specified, the interface will *always* return a JSON array, where each element in the array is a object that represents a table row.

simple example:

```
https://example.com/AdventureWorks/HumanResources/employee?filter=JobTitle~=*Dev*
```

```sql
-- TSQL
SELECT *
FROM [AdventureWorks].[HumanResources].[Employee]
WHERE [JobTitle] LIKE '%Tech%'
```

more complex example:

```
https://example.com/AdventureWorks/HumanResources/employee
?filter=VacationHours>60;CurrentFlag==true;(JobTitle~=*Tech*,JobTitle~=*Dev*)
&columns=BusinessEntityID,JobTitle,HireDate
&orderby=BusinessEntityID
&limit=2
&offset=1
```
```sql
-- TSQL
SELECT
    [BusinessEntityID],
    [JobTitle],
    [HireDate]
FROM [AdventureWorks].[HumanResources].[Employee]
WHERE 
    [VacationHours] > 50
    AND [CurrentFlag] = 1
    AND (
        [JobTitle] LIKE '%Tech%'
        OR [JobTitle] LIKE '%Dev%'
    )
ORDER BY [BusinessEntityID]
OFFSET 10 ROWS 
FETCH NEXT 10 ROWS ONLY;
```

#### Filtering


##### Comparison Operators

delimit each operand with one of these operators. to search for a `IS NULL` or `IS NOT NULL` value, use the `null` key word. 

| URL Operator | SQL Operator | Definition                |
| ------------ | ------------ | ------------------------- |
| ==           | =            | equals                    |
| !=           | =            | not equals                |
| <            | <            | less than                 |
| >            | >            | greater than              |
| <=           | <=           | less than or equal too    |
| >=           | >=           | greater than or equal too |
| ~=           | LIKE         | wild card equals          |
| ==null       | IS NULL      | equals null               |
| !=null       | IS NOT NULL  | not equals null           |

##### Logical Operators

delimit each expression with one of these operators. but keep in mind there is no precedence, so it only make sense to choose one type of operator in each query.

| URL Operator | SQL Operator | Definition       |
| ------------ | ------------ | ---------------- |
| ;            | AND          | exclusive search |
| ,            | OR           | Inclusive search |


##### Wildcards

Used in conjuction with the `~=` operator.

| Wildcard | Definition                                          |
| -------- | --------------------------------------------------- |
| *        | zero or more of any character                       |
| ?        | any one character                                   |
| [a-z]    | a range of possible characters in a single position |
| [abc]    | a list of possible characters in a single position  |


### Inserting and updating Rows

Peck can abstract away inserting or updating data, and simply perform the correct operation on the data you provide it. 
- To perform a insert either provide a JSON object with a new key/value pair that satisfies a unique constraint, or provide a JSON object without satisfying a unique constraint. 
- To update a row provide a JSON object with key/values that match an existing  unique key/values in the table.

#### insert/update one row

*A PUT request should always identify a single resource*, which makes it perfect for providing an interface that will guarantee a single row is either inserted or updated. This interface requires that you identify the resource within the URL fragment. If this interface does not perform a insert or an update it is considered an error. If an error occurs all changes made during the request will be rolled back. This interface will always return a JSON object that describes either the modification to a row or an error. 

`PUT /database/schema/table#column=value&column=value`

```json 
{"key":"value","key":"value"}
```

#### insert/update many rows

Peck can manage batch insert/updates, however for a row to be updated, the JSON object must contain a key/value that matches an existing unique index, otherwise objects sent to this interface will always be inserted. If this interface does not perform a insert or an update it is considered an error. If an error occurs all changes made during the request will be rolled back. This interface will always return a JSON object that describes either the modifications to a row or an error.

`POST /database/schema/table`

```json
[
    {"key":"value","key":"value"},
    {"key":"value","key":"value"}
]
```

### Deleting Data

#### delete a single row

You can delete a single row by specifying its unique index within the URL fragment. If this interface does not delete one row it is considered an error. If an error occurs all changes made during the request will be rolled back. This interface will always return a JSON object that describes either the modifications to a row or an error.

`DELETE /database/schema/table#column=value&column=value`

#### delete many rows

You can delete a single row by specifying its unique index within the URL fragment. If this interface does not delete a single row it is considered an error. If an error occurs all changes made during the request will be rolled back. This interface will always return a JSON object that describes either the modifications to a row or an error.

`DELETE /database/schema/table?filter=expression`

## SSL Certificates
peck will attempt to load two files into memory from its working directory. you can provided these files, or Peck will attempt to create them automatically.
- peck.crt
- peck.key


### Self Signed Certificates
if the `peck.crt` and `peck.key` files cant be found Peck will attempt to generate them automatically. if this fails, you can always generate them yourself.


#### windows

For windows you will need to install `git` which comes with to openssl.

```powershell
& "C:\Program Files\Git\usr\bin\openssl.exe" req -x509 -nodes -days 3650 -newkey rsa:2048 -keyout server.key -out server.crt  -subj '/CN=peck'
```


#### linux or mac

```bash
openssl req -x509 -nodes -days 3650 -newkey rsa:2048 -keyout server.key -out server.crt -subj '/CN=peck'
```
