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
    if results:
        return results[0]
    return {}


def get_row_json(db: str, schema: str, table: str, columns: set, pks: dict) -> str:
    results = get_row(db, schema, table, columns, pks)
    return json.dumps(results, default=json_converter)


def get_data(db: str, query: str):
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


def set_data_json(db, schema: str, table: str, pks: list, json_rows: str) -> int:
    rows = json.loads(json_rows)
    return set_data(db, schema, table, pks, rows)
