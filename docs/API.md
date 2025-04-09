# Cataloging API Documentation

## Overview

This API enables users to register new materials before being created in SAP.

### GET /ping

Check server's health. On a healthy server, it simply returns `200` response header.

#### Example request

```bash
curl --location '{HOST}:{PORT}/ping'
```

#### Example response

- 200

### POST /login

Logs user into the system. The returned access token can be used for authorization purpose when calling most of the endpoints, while
the refresh token can be used to generate new access token if the old one expires.

#### Example request

```bash
curl --location '{HOST}:{PORT}/login' \
--header 'Content-Type: application/json' \
--data-raw '{
    "id": "string,required",
    "password": "string,required"
}'
```

#### Example response

- 200

```json
{
    "accessToken": "string",
    "refreshToken": "string",
    "expiredAt": 0
}
```

- 400, 401, 404, 500

```json
{
    "errorCode": "string"
}
```

### POST /refresh

Generate new access token using a refresh token. Refresh tokens can still be expired although their lifetime is typically much longer than that of access tokens. Once a refresh token expired, users must perform new login.

#### Example request

```bash
curl --location '{HOST}:{PORT}/refresh' \
--header 'Authorization: Bearer string' \
--header 'Content-Type: application/json' \
--data '{
    "id": "string,required"
}'
```

#### Example response

- 200

```json
{
    "accessToken": "string",
    "expiredAt": 0
}
```

- 400, 401, 403, 404, 500

```json
{
    "errorCode": "string"
}
```

### POST /users

Register new user.

#### Example request

```bash
curl --location '{HOST}:{PORT}/users' \
--header 'Content-Type: application/json' \
--data '{
    "id": "string,required",
    "name": "string,required",
    "email": "string,required",
    "password": "string,required"
}'
```

#### Example response

- 201

- 400, 409, 500

```json
{
    "errorCode": "string"
}
```

### GET /users

List existing users with filter, sort criteria and pagination through query parameters. Optional `name` and `isAdmin` parameter are for filtering users based on their name and administrator privilage. Optional `sortBy` parameter accepts `id`, `name`, `email` and `is_admin`, while `isDescending` is either `false` or `true`. Both are for defining sorting criteria. Default sorting criteria is by record's creation time in descending order. Optional `limit` and `page` are for pagination. Default page number and item per page are 1 and 20, respectively. Page number must be greater than 0 and item per page should be between 1 to 20. List users is exclusively available for administrators only.

#### Example request

```bash
curl --location '{HOST}:{PORT}/users?name=string&isAdmin=bool&sortBy=string&isDescending=bool&limit=int&page=int' \
--header 'Authorization: Bearer string'
```

#### Example response

- 200

```json
{
    "data": [
        {
            "id": "string",
            "name": "string",
            "email": "string",
            "isAdmin": false,
            "createdAt": 0,
            "updatedAt": 0
        }
    ],
    "meta": {
        "currentPage": 1,
        "nextPage": null,
        "previousPage": null,
        "totalPages": 1,
        "totalRecords": 1
    }
}
```

- 400, 403, 500

```json
{
    "errorCode": "string"
}
```

### GET /users/{id}

Get user's detail by ID. Only the respective user and administrators can access user's detail.

#### Example request

```bash
curl --location '{HOST}:{PORT}/users/{id}' \
--header 'Authorization: Bearer string'
```

#### Example response

- 200

```json
{
    "data": {
        "id": "string",
        "name": "string",
        "email": "string",
        "isAdmin": false,
        "createdAt": 0,
        "updatedAt": 0
    }
}
```

- 401, 403, 404, 500

```json
{
    "errorCode": "string"
}
```

### PUT /users/{id}

Update user's detail. Only the respective user and administrators can update user's detail.

#### Example request

```bash
curl --location --request PUT '{HOST}:{PORT}/users/{id}' \
--header 'Authorization: Bearer string' \
--header 'Content-Type: application/json' \
--data-raw '{
    "name": "string,required",
    "email": "string,required",
    "password": "string,required"
}'
```

#### Example response

- 204

- 400, 401, 403, 404, 500

```json
{
    "errorCode": "string"
}
```

### DELETE /users/{id}

Delete user's account. Only the respective user and administrators can delete the user's account.

#### Example request

```bash
curl --location --request DELETE '{HOST}:{PORT}/users/{id}' \
--header 'Authorization: Bearer string'
```

### POST /material_types

Create a new material type. It requires an administrator privilege. All material types should be in accordance with SAP Material Management Module Blueprint.

#### Example request

```bash
curl --location '{HOST}:{PORT}/material_types' \
--header 'Authorization: Bearer string' \
--header 'Content-Type: application/json' \
--data '{
    "code": "string,required",
    "description": "string,required",
    "valuationClass": "string,optional"
}'
```

#### Example response

- 204

- 400, 401, 403, 409, 500

```json
{
    "errorCode": "string"
}
```

### POST /material_uoms

Create a new unit of measure. It requires an administrator privilege. All unit of measures should be in accordance with SAP Material Management Module Blueprint.

#### Example request

```bash
curl --location '{HOST}:{PORT}/material_uoms' \
--header 'Authorization: Bearer string' \
--header 'Content-Type: application/json' \
--data '{
    "code": "string,required",
    "description": "string,required"
}'
```

#### Example response

- 204

- 400, 401, 403, 409, 500

```json
{
    "errorCode": "string"
}
```

### POST /material_groups

Create a new material group. It requires an administrator privilege. All material groups should be in accordance with SAP Material Management Module Blueprint.

#### Example request

```bash
curl --location '{HOST}:{PORT}/material_groups' \
--header 'Authorization: Bearer string' \
--header 'Content-Type: application/json' \
--data '{
    "code": "string,required",
    "description": "string,required"
}'
```

#### Example response

- 204

- 400, 401, 403, 409, 500

```json
{
    "errorCode": "string"
}
```

### GET /material_types

List existing material types with filter, sort criteria and pagination through query parameters. Optional `description` parameter are for filtering material types based on their description. Optional `sortBy` parameter accepts `code`, `description` and `val_vlass`, while `isDescending` is either `false` or `true`. Both are for defining sorting criteria. Default sorting criteria is by record's creation time in descending order. Optional `limit` and `page` are for pagination. Default page number and item per page are 1 and 20, respectively. Page number must be greater than 0 and item per page should be between 1 to 20. List material types is available for all users.

#### Example request

```bash
curl --location '{HOST}:{PORT}/material_types?description=string&sortBy=string&isDescending=bool&limit=int&page=int' \
--header 'Authorization: Bearer string'
```

#### Example response

- 200

```json
{
    "data": [
        {
            "code": "string",
            "description": "string",
            "valuationClass": "string",
            "createdAt": 0,
            "updatedAt": 0
        }
    ],
    "meta": {
        "currentPage": 1,
        "nextPage": null,
        "previousPage": null,
        "totalPages": 1,
        "totalRecords": 1
    }
}
```

- 400, 401, 403, 500

```json
{
    "errorCode": "string"
}
```

### GET /material_uoms

List existing unit of measures with filter, sort criteria and pagination through query parameters. Optional `description` parameter are for filtering unit of measures based on their description. Optional `sortBy` parameter accepts `code` and `description`, while `isDescending` is either `false` or `true`. Both are for defining sorting criteria. Default sorting criteria is by record's creation time in descending order. Optional `limit` and `page` are for pagination. Default page number and item per page are 1 and 20, respectively. Page number must be greater than 0 and item per page should be between 1 to 20. List unit of measures is available for all users.

#### Example request

```bash
curl --location '{HOST}:{PORT}/material_uoms?description=string&sortBy=string&isDescending=bool&limit=int&page=int' \
--header 'Authorization: Bearer string'
```

#### Example response

- 200

```json
{
    "data": [
        {
            "code": "string",
            "description": "string",
            "createdAt": 0,
            "updatedAt": 0
        }
    ],
    "meta": {
        "currentPage": 1,
        "nextPage": null,
        "previousPage": null,
        "totalPages": 1,
        "totalRecords": 1
    }
}
```

- 400, 401, 403, 500

```json
{
    "errorCode": "string"
}
```

### GET /material_groups

List existing material groups with filter, sort criteria and pagination through query parameters. Optional `description` parameter are for filtering unit of measures based on their description. Optional `sortBy` parameter accepts `code` and `description`, while `isDescending` is either `false` or `true`. Both are for defining sorting criteria. Default sorting criteria is by record's creation time in descending order. Optional `limit` and `page` are for pagination. Default page number and item per page are 1 and 20, respectively. Page number must be greater than 0 and item per page should be between 1 to 20. List unit of measures is available for all users.

#### Example request

```bash
curl --location '{HOST}:{PORT}/material_groups?description=string&sortBy=string&isDescending=bool&limit=int&page=int' \
--header 'Authorization: Bearer string'
```

#### Example response

- 200

```json
{
    "data": [
        {
            "code": "string",
            "description": "string",
            "createdAt": 0,
            "updatedAt": 0
        }
    ],
    "meta": {
        "currentPage": 1,
        "nextPage": null,
        "previousPage": null,
        "totalPages": 1,
        "totalRecords": 1
    }
}
```

- 400, 401, 403, 500

```json
{
    "errorCode": "string"
}
```

### GET /material_types/{code}

Get material type's detail by code. Get material type's detail is available for all users.

#### Example request

```bash
curl --location '{HOST}:{PORT}/material_types/{code}' \
--header 'Authorization: Bearer string'
```

#### Example response

- 200

```json
{
    "data": {
        "code": "string",
        "description": "string",
        "valuationClass": "string",
        "createdAt": 0,
        "updatedAt": 0
    }
}
```

- 401, 403, 404, 500

```json
{
    "errorCode": "string"
}
```

### GET /material_uoms/{code}

Get unit of measure's detail by code. Get unit of measure's detail is available for all users.

#### Example request

```bash
curl --location '{HOST}:{PORT}/material_uoms/{code}' \
--header 'Authorization: Bearer string'
```

#### Example response

- 200

```json
{
    "data": {
        "code": "string",
        "description": "string",
        "createdAt": 0,
        "updatedAt": 0
    }
}
```

- 401, 403, 404, 500

```json
{
    "errorCode": "string"
}
```

### GET /material_groups/{code}

Get material group's detail by code. Get material group's detail is available for all users.

#### Example request

```bash
curl --location '{HOST}:{PORT}/material_groups/{code}' \
--header 'Authorization: Bearer string'
```

#### Example response

- 200

```json
{
    "data": {
        "code": "string",
        "description": "string",
        "createdAt": 0,
        "updatedAt": 0
    }
}
```

- 401, 403, 404, 500

```json
{
    "errorCode": "string"
}
```

### PUT /material_types/{code}

Update an existing material type by code. This is available for administrators only.

#### Example request

```bash
curl --location --request PUT '{HOST}:{PORT}/material_types/{code}' \
--header 'Authorization: Bearer string' \
--header 'Content-Type: application/json' \
--data '{
    "description": "string,required",
    "valuationClass": "string,optional"
}'
```

#### Example response

- 204

- 400, 401, 403, 404, 500

```json
{
    "errorCode": "string"
}
```

### PUT /material_uoms/{code}

Update an existing unit of measure by code. This is available for administrators only.

#### Example request

```bash
curl --location --request PUT '{HOST}:{PORT}/material_uoms/{code}' \
--header 'Authorization: Bearer string' \
--header 'Content-Type: application/json' \
--data '{
    "description": "string,required"
}'
```

#### Example response

- 204

- 400, 401, 403, 404, 500

```json
{
    "errorCode": "string"
}
```

### PUT /material_groups/{code}

Update an existing material group by code. This is available for administrators only.

#### Example request

```bash
curl --location --request PUT '{HOST}:{PORT}/material_groups/{code}' \
--header 'Authorization: Bearer string' \
--header 'Content-Type: application/json' \
--data '{
    "description": "string,required"
}'
```

#### Example response

- 204

- 400, 401, 403, 404, 500

```json
{
    "errorCode": "string"
}
```

### DELETE /material_types/{code}

Delete an existing material type by code. This is available for administrators only. Normally, delete action on arbritary material type will not return error.

#### Example request

```bash
curl --location --request DELETE '{HOST}:{PORT}/material_types/{code}' \
--header 'Authorization: Bearer string'
```

#### Example response

- 204

- 401, 403, 500

```json
{
    "errorCode": "string"
}
```

### DELETE /material_uoms/{code}

Delete an existing unit of measure by code. This is available for administrators only. Normally, delete action on arbritary unit of measure will not return error.

#### Example request

```bash
curl --location --request DELETE '{HOST}:{PORT}/material_uoms/{code}' \
--header 'Authorization: Bearer string'
```

#### Example response

- 204

- 401, 403, 500

```json
{
    "errorCode": "string"
}
```

### DELETE /material_groups/{code}

Delete an existing material group by code. This is available for administrators only. Normally, delete action on arbritary material group will not return error.

#### Example request

```bash
curl --location --request DELETE '{HOST}:{PORT}/material_groups/{code}' \
--header 'Authorization: Bearer string'
```

#### Example response

- 204

- 401, 403, 500

```json
{
    "errorCode": "string"
}
```