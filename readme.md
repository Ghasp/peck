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
WHERE [JobTitle] LIKE '%Dev%'
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

where `k` is the key and `v` is a value.

##### Comparison Operators

To escape an operator use `\`.

| URL Operator | SQL Operator | Definition                |
| ------------ | ------------ | ------------------------- |
| `k`==`v`     | =            | equals                    |
| `k`!=`v`     | =            | not equals                |
| `k`<`v`      | <            | less than                 |
| `k`>`v`      | >            | greater than              |
| `k`<=`v`     | <=           | less than or equal too    |
| `k`>=`v`     | >=           | greater than or equal too |
| `k`~=`v`     | LIKE         | wild card equals          |
| `k`__        | IS NULL      | equals null               |
| `k`!_        | IS NOT NULL  | not equals null           |

##### Wildcards

Used in conjuction with the `~=` operator. To escape an wildcard use `\`.

| Wildcard | Definition                                          |
| -------- | --------------------------------------------------- |
| *        | zero or more of any character                       |
| ?        | any one character                                   |
| [a-z]    | a range of possible characters in a single position |
| [abc]    | a list of possible characters in a single position  |

##### Ternary Operators

| URL Operator  | SQL Operator | Definition                       |
| ------------- | ------------ | -------------------------------- |
| `k`==`v`<>`v` | BETWEEN      | within a range of the two values |

##### Logical Operators

expressions can be delimited with these operators. Where `e` is a expression.

| URL Operator | SQL Operator | Definition       |
| ------------ | ------------ | ---------------- |
| `e`;`e`      | AND          | exclusive search |
| `e`,`e`      | OR           | Inclusive search |
| (`e`)        | ( )          | Precedence       |



### Inserting and updating Rows

Peck can abstract away inserting or updating data, and simply perform the correct operation on the data you provide it. 

- To perform a insert either provide a JSON object with a new key/value pair that satisfies a unique constraint, or provide a JSON object without satisfying a unique constraint. 
- To update a row provide a JSON object with key/values that match an existing  unique key/values in the table.

#### insert/update one row

*A PUT request should always identify a single resource*, which makes it perfect for providing an interface that will guarantee a single row is either inserted or updated. This interface requires that you identify the resource within the URL fragment. If a request to this interface does not result in a insert or an update it is considered an error. If an error occurs all changes made during the request will be rolled back. This interface will always return a JSON object that describes either the modification to a row or an error. 

`PUT /database/schema/table#column=value&column=value`

```json 
{"key":"value","key":"value"}
```

#### insert/update many rows

Peck can manage batch insert/updates, however for a row to be updated, the JSON object must contain a key/value that matches an existing unique index, otherwise it will be inserted. If a request to this interface does not result in a insert or an update it is considered an error. If an error occurs all changes made during the request will be rolled back. This interface will always return a JSON object that describes either the modifications to a row or an error.

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




## Why Fragments?

Peck uses the URI *path* to identify all the rows in a table e.g. `/database/schema/table`. The URI *query* to filter the rows in the table. And the URI *fragment* to uniquely identify a row in a table.

**Why use the URI fragment to identify a row?**

Accourding to the URI [rfc3986](https://tools.ietf.org/html/rfc3986), fragments are used to identify a "secondary resource" (a row) which can be a subset of the "primary recourse" (a table). Since SQL tables are capable of being uniquely indexed by multiple columns and those columns can be named almost anything, the URI path can not naturally contain all that data. The URI fragment is intended to identify an existing resource within the primary resource and can store arbitrary data about that secondary resource to identify it.

If SQL table enforces a single primary key with a revered column name (like many other nosql databases do) then a row could be referenced with a URL path like `/database/schema/table/id`. unfortunately this is not the case. instead a row in a sql database must be identifies with an array of key/value pairs e.g `/database/schema/table#key=value&key=value`.

**Why not use the URI query?** 

The URI query is not intended to uniquely identify a secondary resouce, it is used to store arbitrary data that further describes the primary resouce. while it could be used to identify a secondary resource, this does not seem to be its *only* purpose. Which is why Peck uses it to filter the rows in a table and not uniquely identify a single row.

**But arn't URI fragments not handled by servers when browsing the web?**

Its important to keep in mind that URI's are used for all sorts of things not just web browsers. So applying web browers rules to a REST API will not always make sense, especially when web browers arent going to be the main clients using the API. but for the record, the fragment part of a URI is usually sent to servers from a web brower, but most servers just choose to ignore it.
