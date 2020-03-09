
# a small query language that will fit in a url query paramter.

comprehensive_example = """

key~=value;((key==value;key!=value),(key>value;key<value;),(key>=value,key=<=value));key__;key!_

"""

# WHERE
#   key LIKE value
#    AND (
#       (key = value AND key != value)
#       OR (key > value AND key < value)
#       OR (key >= value AND key <= value)
#    )
#    AND key is null
#    AND key is not null


class exclusive():
    DELIMITER = ';'


class inclusive():
    DELIMITER = ','


class scope(list):
    def __init__(self, raw: str, *args, **kwargs):
        self = []
        self.raw = raw

    search_type = None
    raw = ''
    OPEN = '('
    CLOSE = ')'


class comparison():
    def __init__(self, raw: str, *args, **kwargs):
        self.raw = raw
    raw = ''
    left = None
    operator = None
    right = None

    # search from right to left for an operator
    def parse(self, raw: str):
        self.raw = raw

        index =  0
        
        while index < len(raw):
            try:
                raw[index]


            finally:
                index += 1



class operator():
    raw = None

class eq(operator):
    SYMBOL = '=='

class ne(operator):
    SYMBOL = '!='

class lt(operator):
    SYMBOL = '<'

class gt(operator):
    SYMBOL = '>'

class le(operator):
    SYMBOL = '<='
    
class ge(operator):
    SYMBOL = '>='



# hunts for closing (), or end of line.
def parse_to_end_of_scope(raw: str) -> str:
    open_count = 0
    for i in range(len(raw)):
        char = raw[i]
        if open_count < 0:
            raise Exception('bad scope')
        if char == scope.OPEN:
            open_count += 1
        if char == scope.CLOSE:
            open_count -= 1
            if open_count == 0:
                return raw[1:i]
    return raw


def parse_to_begining_of_comparison(raw: str) -> str:
    return raw


def parse_scope(raw: str) -> scope:

    current_scope = scope()
    index = 0

    while index < len(raw):
        try:
            char = raw[index]

            if char == scope.OPEN:
                # search for the end of the inner scope
                inner_scope_raw = parse_to_end_of_scope(raw[index:])

                # parse the inner scope and append as a scope element
                inner_scope = parse_scope(inner_scope_raw)
                current_scope.append(inner_scope)

                # set the index to the end of the inner scope
                index = index + len(inner_scope_raw) - 2
                if index < 0:
                    raise Exception('oops!')

                # continue to increment index
                continue

            elif char == inclusive.DELIMITER or char == exclusive.DELIMITER:
                # if this is the first delimiter, define scope type
                # if this is not the first delimiter, check the scope
                if isinstance(current_scope.search_type, inclusive):
                    if char == exclusive.DELIMITER:
                        raise Exception(
                            'Can not mix inclusive and inclusive searches in the same scope.')
                elif isinstance(current_scope.search_type, exclusive):
                    if char == inclusive.DELIMITER:
                        raise Exception(
                            'Can not mix inclusive and inclusive searches in the same scope.')
                else:
                    if char == inclusive.DELIMITER:
                        current_scope.search_type = inclusive()
                    elif char == exclusive.DELIMITER:
                        current_scope.search_type = exclusive()

                # search for the begining of the element
                # parse the selected element as an operator
                # append the operator as an element in scope

        finally:
            index += 1

    # ensure scoping was set correctly
    if current_scope.search_type = None and len(current_scope) > 1:
        raise Exception('search_type was never set correctly.')
    if current_scope.search_type = None and len(current_scope) == 1:
        current_scope.search_type = exclusive()


    return current_scope


def convert(raw: str, depth: int = 0, index: int = 0) -> list:

    expression = None
    scoped = scope()

    cursor = index
    select = index

    while index < len(raw):
        try:

            # find in scopes
            if raw[index] == scope.OPEN:
                cursor = index
                count = 1
                closed = False
                while not closed:
                    select += 1
                    char = raw[select]
                    if char == scope.OPEN:
                        count += 1
                    if char == scope.CLOSE:
                        count -= 1
                        if count < 0:
                            raise Exception(
                                f'To many closing parentheses as index {select}.')
                        if count == 0:
                            closed = True

                scoped = raw[cursor+1:select]
                index = select

            # find and delimit scopes
            if raw[index] == inclusive.DELIMITER or raw[index] == exclusive.DELIMITER:
                if start_expression:
                    start_expression = False
                    if select < len(raw):
                        char = raw[select]
                        if char == exclusive.DELIMITER:
                            expression = exclusive()
                        elif char == inclusive.DELIMITER:
                            expression = inclusive()
                    else:
                        expression = exclusive()
                    expression.append(scoped)
                elif select < len(raw):
                    char = raw[select]
                    if isinstance(expression, exclusive):
                        if char == exclusive.DELIMITER:
                            expression.append(scoped)
                        elif char == inclusive.DELIMITER:
                            raise Exception(
                                'you can not mix an inclusive search in the same scope as exclusive search.')
                    elif isinstance(expression, inclusive):
                        if char == inclusive.DELIMITER:
                            expression.append(scoped)
                        elif char == exclusive.DELIMITER:
                            raise Exception(
                                'you can not mix an exclusive search in the same scope as inclusive search.')
                else:
                    expression.append(scoped)

        finally:
            index += 1
            cursor = index
            select = index
    return expression


print(convert("(1(2)3),(4(5)6),(7(8)9)"))
