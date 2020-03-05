from aiohttp import web, BasicAuth
from db import new_mssql_conn_str, get_row_json, set_data_json
from tsql import to_upsert, escape_string
import json
from util import decode_basic_atuh

routes = web.RouteTableDef()

server = 'localhost'

@routes.get('/{database}/{schema}/{table}')
async def get_table_data(request):
    database = request.match_info.get('database')
    schema = request.match_info.get('schema')
    table = request.match_info.get('table')
    username, password = decode_basic_atuh(request.headers["Authorization"])

    keys = {}
    for key in request.query:
        if str(key).lower() == 'limit':
            continue
        if str(key).lower() == 'offset':
            continue
        keys.update({key: request.query[key]})

    conn_str = new_mssql_conn_str(server, database, username, password)
    results = get_row_json(conn_str, schema, table, [], keys)

    return web.Response(
        text=results,
        content_type='application/json'
    )


@routes.post('/{database}/{schema}/{table}')
async def set_table_data(request):
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
        keys = str(request.query['keys']).lower().split(',')
        if not keys:
            raise "keys parameter can not be empty."

        log['modified'] = set_data_json(conn_str, schema, table, keys, rows)

    except Exception as err:
        log["error"] = str(err)
        return web.Response(
            body=json.dumps(log),
            content_type='application/json'
        )

    return web.Response(body=json.dumps(log))


# @routes.put('/student_transfers/{id}')
# async def put_student_transfers(request):
#    try:
#        ID = request.match_info.get('id')
#        data: dict = await request.json()
#        data.update({"RequestID": ID})
#        upsert_stmt = to_upsert("dbo", "TransferRequests", ["RequestID"], data)
#        print(upsert_stmt)
#        do_change(db_connection, upsert_stmt)
#    except Exception as e:
#        return web.Response(
#            status=web.HTTPBadRequest,
#            body=json.dumps({"Error": e}),
#            content_type='application/json'
#        )
#    return web.Response(status=201)
#

app = web.Application()
app.router.add_routes(routes)

if __name__ == '__main__':
    web.run_app(app)
