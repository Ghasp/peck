

def escape_string(s: str) -> str:
    return s.replace("'", "''")


def to_select(schema: str, table: str, columns: list, wheres: dict, close: bool = True) -> str:
    stmt = ''

    select = ''
    if columns:
        for i in range(len(columns)):
            if i > 0:
                select = select + ", "
            select = select + "[" + columns[i] + "]"
    else:
        select = "*"

    full_table_name = '[{}].[{}]'.format(schema, table)
    stmt = 'SELECT {} FROM {}'.format(select, full_table_name)

    if wheres:
        keys = list(wheres.keys())
        stmt = stmt + " WHERE "
        for i in range(len(keys)):
            column = keys[i]
            value = wheres[column]
            if i > 0:
                stmt = stmt + " AND "
            stmt = stmt + "[" + column + "] = '" + \
                escape_string(str(value)) + "'"

    if close:
        stmt = stmt + ";"

    return stmt


def to_insert(schema: str, table: str, row: dict) -> str:
    keys = list(row.keys())

    columns = ""
    for i in range(len(keys)):
        if i > 0:
            columns = columns + ", "
        columns = columns + "[" + keys[i] + "]"

    values = ""
    for i in range(len(row)):
        key = keys[i]
        if i > 0:
            values = values + ", "
        values = values + "'" + escape_string(str(row[key])) + "'"

    return "INSERT INTO [{}].[{}] ({}) VALUES ({});".format(schema, table, columns, values)


def to_update(schema: str, table: str, row: dict, wheres: dict) -> str:

    full_table_name = '[{}].[{}]'.format(schema, table)
    stmt = "UPDATE {} SET ".format(full_table_name)

    updates = list(row.keys())
    for i in range(len(updates)):
        column = updates[i]
        if i > 0:
            stmt = stmt + ", "
        stmt = stmt + "[" + column + "] = '" + \
            escape_string(str(row[column])) + "'"

    if wheres:
        keys = list(wheres.keys())
        stmt = stmt + " WHERE "
        for i in range(len(keys)):
            column = keys[i]
            value = wheres[column]
            if i > 0:
                stmt = stmt + " AND "
            stmt = stmt + "[" + column + "] = '" + \
                escape_string(str(value)) + "'"

    stmt = stmt + ';'
    return stmt


def to_delete(schema: str, table: str, wheres: dict) -> str:
    full_table_name = '[{}].[{}]'.format(schema, table)
    stmt = "DELETE FROM {}".format(full_table_name)
    if wheres:
        keys = list(wheres.keys())
        stmt = stmt + " WHERE "
        for i in range(len(keys)):
            column = keys[i]
            value = wheres[column]
            if i > 0:
                stmt = stmt + " AND "
            stmt = stmt + "[" + column + "] = '" + \
                escape_string(str(value)) + "'"
    stmt = stmt + ';'
    return stmt


def to_upsert(schema: str, table: str, pks: list, row: dict) -> str:

    #columns = list(row.keys())

    wheres = {}
    for column in row:
        if str(column) in pks:
            wheres.update({column: row[column]})

    if not wheres:
        raise ValueError("please provide '" + "', '".join(pks) + "' values.")

    select = to_select(schema, table, [], wheres, False)
    insert = to_insert(schema, table, row)
    update = to_update(schema, table, row, wheres)
    stmt = 'IF NOT EXISTS ({}) BEGIN {} END ELSE BEGIN {} END'.format(
        select, insert, update)
    return stmt