from aiohttp import web
from db import (
    new_mssql_conn_str,
    get_row_json,
    get_rows_json,
    set_data_json,
    get_json_data,
    get_table_primary_keys,
    do_sql,
    delete_row
)
from tsql import to_upsert, escape_string
from util import decode_basic_atuh, test_path, request_basic_auth
import json
import ssl
import subprocess
import os
import sys
import pyodbc
import re


routes = web.RouteTableDef()
server = os.environ['PECK_DB_SERVER']


@routes.get("/")
async def home(request):
    html = ''
    with open("readme.html", encoding="utf8") as file:
        html = file.read()

    return web.Response(
        body=html,
        content_type='text/html'
    )


# provides a interface for sending sql querys to retrieve json data.
# any changes to the database will not stick and will be rolled back.
@routes.post('/{database}/query')
async def query_database(request):
    log = {
        "error": ""
    }

    try:
        database = request.match_info.get('database')
        username, password = decode_basic_atuh(
            request.headers["Authorization"])
        query = await request.text()

        if not query:
            raise Exception(
                'please provide a sql query in the http request body.')

        conn_str = new_mssql_conn_str(server, database, username, password)
        results = get_json_data(conn_str, query)
    except KeyError as err:
        return web.Response(
            status=401,
            headers={
                "WWW-Authenticate": "Basic realm=\"please provide a valid sql username and password\"."
            }
        )
    except Exception as err:
        p = re.compile(r'^.*Login failed.*$')
        if p.match(str(err)):
            return web.Response(
                status=401,
                headers={
                    "WWW-Authenticate": "Basic realm=\"please provide a valid sql username and password\"."
                }
            )
        log["error"] = str(err)
        return web.Response(
            body=json.dumps(log),
            content_type='application/json',
            status=400
        )

    return web.Response(
        text=results,
        content_type='application/json'
    )

# provides a interface for executing changes to data with sql.
# if an error is encountered all changes will roll back.
@routes.post('/{database}/modify')
async def exec_database(request):
    log = {
        "modified": 0,
        "error": ""
    }

    try:

        database = request.match_info.get('database')
        username, password = decode_basic_atuh(
            request.headers["Authorization"])
        stmt = await request.text()

        if not stmt:
            raise Exception(
                'please provide a sql stmt in the http request body.')

        conn_str = new_mssql_conn_str(server, database, username, password)
        log["modified"] = do_sql(conn_str, stmt)
    except KeyError as err:
        return web.Response(
            status=401,
            headers={
                "WWW-Authenticate": "Basic realm=\"please provide a valid sql username and password\"."
            }
        )
    except Exception as err:
        p = re.compile(r'^.*Login failed.*$')
        if p.match(str(err)):
            return web.Response(
                status=401,
                headers={
                    "WWW-Authenticate": "Basic realm=\"please provide a valid sql username and password\"."
                }
            )
        log["error"] = str(err)
        return web.Response(
            body=json.dumps(log),
            content_type='application/json',
            status=400
        )

    return web.Response(
        body=json.dumps(log),
        content_type='application/json'
    )


# get a single row.
@routes.get('/{database}/{schema}/{table}')
async def get_row(request):
    log = {
        "error": ""
    }

    try:
        database = request.match_info.get('database')
        schema = request.match_info.get('schema')
        table = request.match_info.get('table')
        username, password = decode_basic_atuh(
            request.headers["Authorization"])

        keys = {}
        for key in request.query:
            keys.update({key: request.query[key]})

        if len(keys) < 1:
            raise Exception(
                'Please specify a rows unique index in the URI query, e.g. ?column=value&column=value. The URL query must uniquely identify a single row.')

        conn_str = new_mssql_conn_str(server, database, username, password)
        results = get_row_json(conn_str, schema, table, [], keys)
    except KeyError as err:
        return web.Response(
            status=401,
            headers={
                "WWW-Authenticate": "Basic realm=\"please provide a valid sql username and password\"."
            }
        )
    except Exception as err:
        p = re.compile(r'^.*Login failed.*$')
        if p.match(str(err)):
            return web.Response(
                status=401,
                headers={
                    "WWW-Authenticate": "Basic realm=\"please provide a valid sql username and password\"."
                }
            )
        log["error"] = str(err)
        return web.Response(
            body=json.dumps(log),
            content_type='application/json',
            status=400
        )

    return web.Response(
        text=results,
        content_type='application/json'
    )

# this is useful for getting all rows with a specified foreign key.
# this will always return an array.
@routes.get('/{database}/{schema}/{table}/join')
async def get_row_by_foriegn_key(request):
    log = {
        "error": ""
    }

    try:

        database = request.match_info.get('database')
        schema = request.match_info.get('schema')
        table = request.match_info.get('table')
        username, password = decode_basic_atuh(
            request.headers["Authorization"])

        keys = {}
        for key in request.query:
            keys.update({key: request.query[key]})

        if len(keys) < 1:
            raise Exception(
                'please specify a foreign key index within the URI query')

        conn_str = new_mssql_conn_str(server, database, username, password)
        results = get_rows_json(conn_str, schema, table, [], keys)

    except KeyError as err:
        return web.Response(
            status=401,
            headers={
                "WWW-Authenticate": "Basic realm=\"please provide a valid sql username and password\"."
            }
        )
    except Exception as err:
        p = re.compile(r'^.*Login failed.*$')
        if p.match(str(err)):
            return web.Response(
                status=401,
                headers={
                    "WWW-Authenticate": "Basic realm=\"please provide a valid sql username and password\"."
                }
            )
        log["error"] = str(err)
        return web.Response(
            body=json.dumps(log),
            content_type='application/json',
            status=400
        )

    return web.Response(
        text=results,
        content_type='application/json'
    )


@routes.put('/{database}/{schema}/{table}')
async def set_table_row(request):
    log = {
        "modified": 0,
        "error": ""
    }

    try:
        rows = await request.text()
        database = request.match_info.get('database')
        schema = request.match_info.get('schema')
        table = request.match_info.get('table')
        username, password = decode_basic_atuh(
            request.headers["Authorization"])

        keys = {}
        for key in request.query:
            keys.update({key: request.query[key]})

        conn_str = new_mssql_conn_str(server, database, username, password)
        log['modified'] = set_data_json(conn_str, schema, table, keys, rows)

    except Exception as err:
        log["error"] = str(err)
        return web.Response(
            body=json.dumps(log),
            content_type='application/json',
            status=400
        )

    return web.Response(body=json.dumps(log))


@routes.post('/{database}/{schema}/{table}')
async def set_table_rows(request):
    log = {
        "modified": 0,
        "error": ""
    }

    try:
        rows = await request.text()

        database = request.match_info.get('database')
        schema = request.match_info.get('schema')
        table = request.match_info.get('table')
        username, password = decode_basic_atuh(
            request.headers["Authorization"])

        conn_str = new_mssql_conn_str(server, database, username, password)
        keys = request.query.getall('key', None)
        if not keys:
            pks = get_table_primary_keys(conn_str, schema, table)
            if not pks:
                raise Exception(
                    "the specified table does not have a primary key column.")
            keys = pks

        log['modified'] = set_data_json(conn_str, schema, table, keys, rows)

    except KeyError as err:
        return web.Response(
            status=401,
            headers={
                "WWW-Authenticate": "Basic realm=\"please provide a valid sql username and password\"."
            }
        )
    except Exception as err:
        p = re.compile(r'^.*Login failed.*$')
        if p.match(str(err)):
            return web.Response(
                status=401,
                headers={
                    "WWW-Authenticate": "Basic realm=\"please provide a valid sql username and password\"."
                }
            )
        log["error"] = str(err)
        return web.Response(
            body=json.dumps(log),
            content_type='application/json',
            status=400
        )

    return web.Response(body=json.dumps(log))


@routes.delete('/{database}/{schema}/{table}')  # ?column=value
async def delete_table_row(request):
    log = {
        "deleted": 0,
        "error": ""
    }

    try:

        database = request.match_info.get('database')
        schema = request.match_info.get('schema')
        table = request.match_info.get('table')
        username, password = decode_basic_atuh(
            request.headers["Authorization"])

        conn_str = new_mssql_conn_str(server, database, username, password)

        wheres = {}
        for key in request.query:
            wheres.update({key: request.query[key]})

        if not wheres:
            raise Exception(
                "you must specify the primary keys in the url query parameters")

        deleted = delete_row(conn_str, schema, table, wheres)

        if deleted:
            log['deleted'] = 1
        else:
            log["error"] = "row does not exist."
            return web.Response(
                body=json.dumps(log),
                content_type='application/json',
                status=400
            )

    except KeyError as err:
        return web.Response(
            status=401,
            headers={
                "WWW-Authenticate": "Basic realm=\"please provide a valid sql username and password\"."
            }
        )
    except Exception as err:
        p = re.compile(r'^.*Login failed.*$')
        if p.match(str(err)):
            return web.Response(
                status=401,
                headers={
                    "WWW-Authenticate": "Basic realm=\"please provide a valid sql username and password\"."
                }
            )
        log["error"] = str(err)
        return web.Response(
            body=json.dumps(log),
            content_type='application/json',
            status=400
        )

    return web.Response(body=json.dumps(log))


if __name__ == '__main__':

    # create self signed cert
    # if not test_path('peck.crt') or not test_path('peck.key'):
    #    print("Attempting to create SSL certs....")
    #    if sys.platform.startswith('win32'):
    #        terminal = "powershell.exe"
    #        openssl_path = os.path.join(
    #            "C:\\", "Program Files", "Git", "usr", "bin", "openssl.exe")
    #    else:
    #        terminal = "/bin/bash"
    #        openssl_path = os.path.join("openssl.exe")
    #
    #    cmd = subprocess.Popen([
    #        terminal,
    #        f"& '{openssl_path}' req -x509 -nodes -days 3650 -newkey rsa:2048 -keyout peck.key -out peck.crt  -subj '/CN=peck'"
    #    ],
    #        stdout=subprocess.PIPE,
    #        universal_newlines=True
    #    )
    #    for line in cmd.stdout:
    #        print(line.strip())
    #
    #ssl_context = ssl.create_default_context(ssl.Purpose.CLIENT_AUTH)
    #ssl_context.load_cert_chain('peck.crt', 'peck.key')

    print("Starting HTTP Server...")
    for route in routes:
        print(route.method + " " + route.path)

    app = web.Application()
    app.router.add_routes(routes)
    #web.run_app(app, ssl_context=ssl_context, port=443)
    web.run_app(app, port=80)
