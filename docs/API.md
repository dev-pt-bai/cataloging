# Cataloging API Documentation

## Overview

This API enables users to register new materials before being created in SAP.

### GET /ping

Check server's health. On a healthy server, it simply returns `200` response header.

#### Example request

```bash
curl --location '[host]:[port]/ping'
```

#### Example response

- 200

### POST /assets

Upload file to Microsoft Onedrive. Only pdf, jpg (or jpeg) and png files are acceptable. There is a limit for maximum allowable file size and it can be configured as part of the environment variables. The uploaded files are meant as attachment for materials. It could be material's datasheet, catalogue, drawing, picture or else.

#### Example request

```bash
curl --location '[host]:[port]/assets' \
--header 'Authorization: Bearer [token]' \
--form 'file=@"path/to/file"'
```

#### Example response

- 201

- 400, 401, 403, 415, 502

```json
{
    "errorCode": "string",
    "requestID": "string"
}
```

### DELETE /assets/{id}

Delete file (permanently) from Microsoft Onedrive. Only the file uploader and administrator can delete an asset.

#### Example request

```bash
curl --location --request DELETE '[host]:[port]/assets/{id}' \
--header 'Authorization: Bearer [token]'
```

#### Example response

- 204

- 401, 403, 502

```json
{
    "errorCode": "string",
    "requestID": "string"
}
```

### GET /settings/msgraph

Get an URL for login to Microsoft with delegated authentication flow. The administrator is required to login via the provided URL, each time the application is started. After that, the application will periodically renew the access token to Microsoft Graph API in the background. Without this login, features that incorporates calling the Microsoft Graph API, such as sending email, upload and delete file will not be successful.

#### Example request

```bash
curl --location '[host]:[port]/settings/msgraph' \
--header 'Authorization: Bearer [token]'
```

#### Example response

- 200

```json
{
    "url": "string"
}
```

- 401, 403, 422

```json
{
    "errorCode": "string",
    "requestID": "string"
}
```

### GET /settings/msgraph/auth

Serves solely as the redirect URI in the delegated authentication flow of Microsoft Graph API. It has no use outside this context. It parses the query parameter with the key `code`, returned after the administrator successfully login to Microsoft. The code is then used to fetch the first access token of Microsoft Graph API.

#### Example request

```bash
curl --location 'localhost:8002/settings/msgraph/auth?code=string&error=string&error_description=string&state=string'
```

#### Example response

- 204

- 422, 500, 502

```json
{
    "errorCode": "string",
    "requestID": "string"
}
```

### POST /login

Logs user into the system. The returned access token can be used for authorization purpose when calling most of the endpoints, while
the refresh token can be used to generate new access token if the old one expires.

#### Example request

```bash
curl --location '[host]:[port]/login' \
--header 'Content-Type: application/json' \
--data '{
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
    "errorCode": "string",
    "requestID": "string"
}
```

### POST /refresh

Generate new access token using a refresh token. Refresh tokens can still be expired although their lifetime is typically much longer than that of access tokens. Once a refresh token expired, users must perform new login.

#### Example request

```bash
curl --location '[host]:[port]/refresh' \
--header 'Authorization: Bearer [token]' \
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
    "errorCode": "string",
    "requestID": "string"
}
```

### POST /users

Register new user.

#### Example request

```bash
curl --location '[host]:[port]/users' \
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
    "errorCode": "string",
    "requestID": "string"
}
```

### GET /users

List existing users with filter, sort criteria and pagination through query parameters. Optional `name`, `role` and `isVerified` parameter are for filtering users based on their name and administrator privilage. Optional `sortBy` parameter accepts `id`, `name`, `email` and `role`, while `isDescending` is either `false` or `true`. Both are for defining sorting criteria. Default sorting criteria is by record's creation time in descending order. Optional `limit` and `page` are for pagination. Default page number and item per page are 1 and 20, respectively. Page number must be greater than 0 and item per page should be between 1 to 20. List users is exclusively available for administrators only.

#### Example request

```bash
curl --location '[host]:[port]/users?name=string&role=int&isVerified=bool&sortBy=string&isDescending=bool&limit=int&page=int' \
--header 'Authorization: Bearer [token]'
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
            "role": "string",
            "isVerified": false,
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
    "errorCode": "string",
    "requestID": "string"
}
```

### GET /users/{id}

Get user's detail by ID. Only the respective user and administrators can access user's detail.

#### Example request

```bash
curl --location '[host]:[port]/users/{id}' \
--header 'Authorization: Bearer [token]'
```

#### Example response

- 200

```json
{
    "data": {
        "id": "string",
        "name": "string",
        "email": "string",
        "role": "string",
        "isVerified": false,
        "createdAt": 0,
        "updatedAt": 0
    }
}
```

- 401, 403, 404, 500

```json
{
    "errorCode": "string",
    "requestID": "string"
}
```

### PUT /users/{id}

Update user's detail. Only the respective user and administrators can update user's detail.

#### Example request

```bash
curl --location --request PUT '[host]:[port]/users/{id}' \
--header 'Authorization: Bearer [token]' \
--header 'Content-Type: application/json' \
--data '{
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
    "errorCode": "string",
    "requestID": "string"
}
```

### DELETE /users/{id}

Delete user's account. Only the respective user and administrators can delete the user's account.

#### Example request

```bash
curl --location --request DELETE '[host]:[port]/users/{id}' \
--header 'Authorization: Bearer [token]'
```

#### Example response

- 204

- 401, 403, 500

```json
{
    "errorCode": "string",
    "requestID": "string"
}
```

### GET /users/{id}/verification

Send an email with verification code to user. To be able to create a material request, user must be verified. An email can be used by multiple users, in which each user will get different verification code. A verification code typically lasts for 5 (five) minutes before it becomes expired. The code should be sent back to the server through `POST /users/{id}/verification` within this time limit.

#### Example request

```bash
curl --location '[host]:[port]/users/{id}/verification' \
--header 'Authorization: Bearer [token]'
```

#### Example response

- 202

- 401, 403, 404, 409, 500, 502

```json
{
    "errorCode": "string",
    "requestID": "string"
}
```

### POST /users/{id}/verification

Verify the user by sending a verification code which has been sent previously from `GET /users/{id}/verification`. In a successful attempt, it will return a new access token which marks that the user has been verified. Verification should only be carried out once. Re-verifying the already-verified user will result in an error. However, when the user's email is changed via an update, it needs to be re-verified.

#### Example request

```bash
curl --location '[host]:[port]/users/{id}/verification' \
--header 'Authorization: Bearer [token]' \
--header 'Content-Type: application/json' \
--data '{
    "code": "string"
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
    "errorCode": "string",
    "requestID": "string"
}
```

### POST /material_types

Create a new material type. It requires an administrator privilege. All material types should be in accordance with SAP Material Management Module Blueprint.

#### Example request

```bash
curl --location '[host]:[port]/material_types' \
--header 'Authorization: Bearer [token]' \
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
    "errorCode": "string",
    "requestID": "string"
}
```

### POST /material_uoms

Create a new unit of measure. It requires an administrator privilege. All unit of measures should be in accordance with SAP Material Management Module Blueprint.

#### Example request

```bash
curl --location '[host]:[port]/material_uoms' \
--header 'Authorization: Bearer [token]' \
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
    "errorCode": "string",
    "requestID": "string"
}
```

### POST /material_groups

Create a new material group. It requires an administrator privilege. All material groups should be in accordance with SAP Material Management Module Blueprint.

#### Example request

```bash
curl --location '[host]:[port]/plants' \
--header 'Authorization: Bearer [token]' \
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
    "errorCode": "string",
    "requestID": "string"
}
```

### POST /plants

Create a new plants. It requires an administrator privilege. All plants should be in accordance with SAP Material Management Module Blueprint.

#### Example request

```bash
curl --location '[host]:[port]/material_groups' \
--header 'Authorization: Bearer [token]' \
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
    "errorCode": "string",
    "requestID": "string"
}
```

### GET /material_types

List existing material types with filter, sort criteria and pagination through query parameters. Optional `description` parameter are for filtering material types based on their description. Optional `sortBy` parameter accepts `code`, `description` and `val_vlass`, while `isDescending` is either `false` or `true`. Both are for defining sorting criteria. Default sorting criteria is by record's creation time in descending order. Optional `limit` and `page` are for pagination. Default page number and item per page are 1 and 20, respectively. Page number must be greater than 0 and item per page should be between 1 to 20. List material types is available for all users.

#### Example request

```bash
curl --location '[host]:[port]/material_types?description=string&sortBy=string&isDescending=bool&limit=int&page=int' \
--header 'Authorization: Bearer [token]'
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
    "errorCode": "string",
    "requestID": "string"
}
```

### GET /material_uoms

List existing unit of measures with filter, sort criteria and pagination through query parameters. Optional `description` parameter are for filtering unit of measures based on their description. Optional `sortBy` parameter accepts `code` and `description`, while `isDescending` is either `false` or `true`. Both are for defining sorting criteria. Default sorting criteria is by record's creation time in descending order. Optional `limit` and `page` are for pagination. Default page number and item per page are 1 and 20, respectively. Page number must be greater than 0 and item per page should be between 1 to 20. List unit of measures is available for all users.

#### Example request

```bash
curl --location '[host]:[port]/material_uoms?description=string&sortBy=string&isDescending=bool&limit=int&page=int' \
--header 'Authorization: Bearer [token]'
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
    "errorCode": "string",
    "requestID": "string"
}
```

### GET /material_groups

List existing material groups with filter, sort criteria and pagination through query parameters. Optional `description` parameter are for filtering unit of measures based on their description. Optional `sortBy` parameter accepts `code` and `description`, while `isDescending` is either `false` or `true`. Both are for defining sorting criteria. Default sorting criteria is by record's creation time in descending order. Optional `limit` and `page` are for pagination. Default page number and item per page are 1 and 20, respectively. Page number must be greater than 0 and item per page should be between 1 to 20. List unit of measures is available for all users.

#### Example request

```bash
curl --location '[host]:[port]/material_groups?description=string&sortBy=string&isDescending=bool&limit=int&page=int' \
--header 'Authorization: Bearer [token]'
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
    "errorCode": "string",
    "requestID": "string"
}
```

### GET /material_types/{code}

Get material type's detail by code. Get material type's detail is available for all users.

#### Example request

```bash
curl --location '[host]:[port]/material_types/{code}' \
--header 'Authorization: Bearer [token]'
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
    "errorCode": "string",
    "requestID": "string"
}
```

### GET /material_uoms/{code}

Get unit of measure's detail by code. Get unit of measure's detail is available for all users.

#### Example request

```bash
curl --location '[host]:[port]/material_uoms/{code}' \
--header 'Authorization: Bearer [token]'
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
    "errorCode": "string",
    "requestID": "string"
}
```

### GET /material_groups/{code}

Get material group's detail by code. Get material group's detail is available for all users.

#### Example request

```bash
curl --location '[host]:[port]/material_groups/{code}' \
--header 'Authorization: Bearer [token]'
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
    "errorCode": "string",
    "requestID": "string"
}
```

### PUT /material_types/{code}

Update an existing material type by code. This is available for administrators only.

#### Example request

```bash
curl --location --request PUT '[host]:[port]/material_types/{code}' \
--header 'Authorization: Bearer [token]' \
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
    "errorCode": "string",
    "requestID": "string"
}
```

### PUT /material_uoms/{code}

Update an existing unit of measure by code. This is available for administrators only.

#### Example request

```bash
curl --location --request PUT '[host]:[port]/material_uoms/{code}' \
--header 'Authorization: Bearer [token]' \
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
    "errorCode": "string",
    "requestID": "string"
}
```

### PUT /material_groups/{code}

Update an existing material group by code. This is available for administrators only.

#### Example request

```bash
curl --location --request PUT '[host]:[port]/material_groups/{code}' \
--header 'Authorization: Bearer [token]' \
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
    "errorCode": "string",
    "requestID": "string"
}
```

### DELETE /material_types/{code}

Delete an existing material type by code. This is available for administrators only. Normally, delete action on arbritary material type will not return error.

#### Example request

```bash
curl --location --request DELETE '[host]:[port]/material_types/{code}' \
--header 'Authorization: Bearer [token]'
```

#### Example response

- 204

- 401, 403, 500

```json
{
    "errorCode": "string",
    "requestID": "string"
}
```

### DELETE /material_uoms/{code}

Delete an existing unit of measure by code. This is available for administrators only. Normally, delete action on arbritary unit of measure will not return error.

#### Example request

```bash
curl --location --request DELETE '[host]:[port]/material_uoms/{code}' \
--header 'Authorization: Bearer [token]'
```

#### Example response

- 204

- 401, 403, 500

```json
{
    "errorCode": "string",
    "requestID": "string"
}
```

### DELETE /material_groups/{code}

Delete an existing material group by code. This is available for administrators only. Normally, delete action on arbritary material group will not return error.

#### Example request

```bash
curl --location --request DELETE '[host]:[port]/material_groups/{code}' \
--header 'Authorization: Bearer [token]'
```

#### Example response

- 204

- 401, 403, 500

```json
{
    "errorCode": "string",
    "requestID": "string"
}
```