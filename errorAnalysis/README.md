ErrorAnalysis

## Introduction
    ErrorAnalysis provides error server and uses 'postgres' as backend db.
    Error server works as a http server, receiving error request from clients and save it into db.Then error server
provides restful http api to CRUD error detail.

## Strategy
    Each error sent by client will generate a keyword in error server.When two pieces of error have the same keyword,
they will not show/save as two record, it will only increase field 'times' for a single error record and save the first
error request param.
    If you want to tell two error records, It advised to using `/error/detail/` replacing `/error/`.
    Why this?
    Most kind of error might occur only once or less by 5 times. This kind of error doesn't happen frequently and hard to
eliminate.So it's ok to have this kind of error and do nothing. We only focus errors with same keyword which happen more
than 5 times, or more times as you think ok rather than each error we pay equal attention.
    In which case you should use `/error/detail/`?
    You want to log the users who the error influences.In this case, `/error/` only record the first request but victims
afterwards will lose their information. `/error/detail/` will record each error.

## Example

#### /error/
ClientA sent an error to `POST /error/`:

```json
{
    "message":"/home/web/projects/project_a/main.go: 129 | redis 'nil return'",
    "request": {
        "user_id": 1,
        "timestamp": "2001-01-01 15:01:01"
    }
}
```

Then got thrown error by `GET /error/`:

```json
{
    "count": 1,
    "data": [
        {
            "id": 1,
            "message": "/home/web/projects/project_a/main.go: 129 | redis 'nil return'",
            "keyword": "/1|R'",
            "times": 1,
            "created_at": "2019-07-12T09:57:04.543586Z",
            "created_date": "2019-07-12T00:00:00Z",
            "updated_at": "2019-07-12T09:57:04.543586Z",
            "request": {
                "user_id": 1,
                "timestamp": "2001-01-01 15:01:01"
            }
        },

    "message": "success"
}
```

or by `GET /error/:id/`:

```json
{
   "data": {
           "id": 1,
           "message": "/home/web/projects/project_a/main.go: 129 | redis 'nil return'",
           "keyword": "/1|R'",
           "times": 1,
           "created_at": "2019-07-12T09:57:04.543586Z",
           "created_date": "2019-07-12T00:00:00Z",
           "updated_at": "2019-07-12T09:57:04.543586Z",
           "request": {
               "user_id": 1,
               "timestamp": "2001-01-01 15:01:01"
           }
   }
}
```

After an error is handled, you can delete it through `DELETE /error/:id/`, or delete-all by `POST /error/delete-all/`

#### /error/detail/
