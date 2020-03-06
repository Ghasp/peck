import pyodbc
import json
import datetime
import tsql


def new_mssql_conn_str(server: str, database: str, username: str, password: str) -> str:
    return f'DRIVER={{SQL Server}};SERVER={server};DATABASE={database};UID={username};PWD={password}'


def json_converter(o):
    if isinstance(o, datetime.datetime):
        return o.isoformat()


def get_row(db: str, schema: str, table: str, columns: set, pks: dict) -> dict:
    query = tsql.to_select(schema, table, columns, pks)
    results = get_data(db, query)
    
    if len(results) > 1:
        raise Exception("Query returned more then one row. Please provide a query that uniquely identifies a single row.")
    elif len(results) == 1:
        return results[0]

    raise Exception("Query did not return a row. Please provide a query that uniquely identifies a single row.")
    


def get_row_json(db: str, schema: str, table: str, columns: set, pks: dict) -> str:
    results = get_row(db, schema, table, columns, pks)
    return json.dumps(results, default=json_converter)


def get_rows(db: str, schema: str, table: str, columns: set, pks: dict):
    query = tsql.to_select(schema, table, columns, pks)
    results = get_data(db, query)
    if results:
        return results
    return []


def get_rows_json(db: str, schema: str, table: str, columns: set, pks: dict) -> str:
    results = get_rows(db, schema, table, columns, pks)
    return json.dumps(results, default=json_converter)


def get_data(db: str, query: str) -> list:
    cnxn = pyodbc.connect(db)
    cursor = cnxn.cursor()
    cursor.execute(query)
    columns = []
    for i in range(len(cursor.description)):
        columns.append(cursor.description[i][0])

    results = []
    while True:

        row = cursor.fetchone()
        if not row:
            break

        result = {}
        for i in range(len(columns)):
            result[columns[i]] = row[i]
        results.append(result)

    cnxn.rollback()
    cnxn.close()
    return results


def get_json_data(connection: pyodbc.Connection, query):
    results = get_data(connection, query)
    return json.dumps(results, default=json_converter)


def do_upsert(cursor: pyodbc.Cursor, schema: str, table: str, pks: list, row: dict) -> int:
    if not isinstance(row, dict):
        raise TypeError("data must be a list of dictionaries")

    stmt = tsql.to_upsert(schema, table, pks, row)
    count = cursor.execute(stmt).rowcount
    return count


def set_data(db, schema: str, table: str, pks: list, rows: list) -> int:
    cnxn = pyodbc.connect(db, autocommit=False)
    cursor = cnxn.cursor()
    count = 0
    try:
        if isinstance(rows, list):
            for row in rows:
                count += do_upsert(cursor, schema, table, pks, row)
        elif isinstance(rows, dict):
            row = rows
            count += do_upsert(cursor, schema, table, pks, row)
        else:
            raise TypeError("rows must be of type list or dict.")
    except Exception as err:
        cnxn.rollback()
        raise err

    cnxn.commit()
    cnxn.close()
    return count


def set_data_json(db, schema: str, table: str, pks: dict, json_rows: str) -> int:
    rows = json.loads(json_rows)

    pk_names = []
    if isinstance(pks, list):
        pk_names = pks

    elif isinstance(pks, dict):
        for pk in pks:
            rows.update({pk: pks[pk]})
            pk_names.append(pk)
    else:
        raise Exception("pks must be of type dict or list")

    return set_data(db, schema, table, pk_names, rows)


def get_table_primary_keys(db, schema: str, table: str) -> list:
    query = """
        select col.[name] as pk
        from (select * from sys.tables where schema_name(schema_id) = '{}' and [name] = '{}') as tab
            inner join sys.indexes pk
                on tab.object_id = pk.object_id 
                and pk.is_primary_key = 1
            inner join sys.index_columns ic
                on ic.object_id = pk.object_id
                and ic.index_id = pk.index_id
            inner join sys.columns col
                on pk.object_id = col.object_id
                and col.column_id = ic.column_id
        order by schema_name(tab.schema_id),
            pk.[name],
            ic.index_column_id
    """.format(schema, table)
    results = get_data(db, query)

    if not results:
        return []

    pks = []
    for row in results:
        pks.append(row["pk"])

    return pks


# do_sql performs table modifications and rollsback changes if not successful.
def do_sql(db: str, sql: str) -> int:
    cnxn = pyodbc.connect(db, autocommit=False)
    cursor = cnxn.cursor()
    count = 0
    try:
        count = cursor.execute(sql).rowcount
    except Exception as err:
        cnxn.rollback()
        raise err
    cnxn.commit()
    cnxn.close()
    return count


# is like do_sql but only allows a single row to be modified/delete.
def do_sql_one(db: str, sql: str) -> bool:
    cnxn = pyodbc.connect(db, autocommit=False)
    cursor = cnxn.cursor()
    count = 0
    try:
        count += cursor.execute(sql).rowcount
    except Exception as err:
        cnxn.rollback()
        raise err

    if count == 1:
        cnxn.commit()
    elif count == 0:
        cnxn.rollback()
    elif count > 1:
        cnxn.rollback()
        raise Exception("This sql statement deleted more then one row. changes have been rolled back. please provide a query that only deletes one row.")
    
    cnxn.close()
    return count == 1



def delete_row(db, schema: str, table: str, pks: dict) -> bool:
    stmt = tsql.to_delete(schema,table,pks)
    return do_sql_one(db, stmt)