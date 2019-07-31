# Web API

## Policies

The policy service has a `/policies` endpoint that determines the set of policies that are applicable to
a given submission.  Note:  The results may be dependent on _who_ submits the request.  For example, if
somebody from JHU invokes the policies endpoint, a general "policy for JHU employees" may be included in the results.

### Policies Request

`GET /policy-service/policies?submission=${SUBMISSION_URI}`

or (with encoded submission=${SUBMISSION_URI}))

```HTTP
POST /policy-service/policies
Content-Type application/x-www-form-urlencoded
```

### Policies Response

The response is a list of URIs to Policy resources, decorated with a `type` property:

```json
[
 {
   "id": "http://pass.local:8080/fcrepo/rest/policies/2d/...",
   "type": "funder"
 },
 {
   "id": "http://pass.local:8080/fcrepo/rest/policies/63/...",
   "type": "institution"
 }
]
```

## Repositories

The policy service has a `/repositories` endpoint that, for a given submission, calculates the repositories that may be
deposited into in order to satisfy any applicable policies for that submission.

### Repositories Request

GET `/policy-service/repositories?submission=${SUBMISSION_URI}`
or, with urlencoded (with encoded submission=${SUBMISSION_URI}) as the body:

```HTTP
POST /policy-service/repositories
Content-Type: application/x-www-form-urlencoded
```

### Repositories Response

Response an application/json document that lists repositories sorted into buckets, as follows:

```json
{
   "required": [
       {
           "url": "http://pass.local/fcrepo/rest/repositories/1",
           "selected": true
       }
   ],
   "one-of": [
       [
           {
               "url": "http://pass.local/fcrepo/rest/repositories/2",
               "selected": true
           },
           {
               "url": "http://pass.local/fcrepo/rest/repositories/3",
               "selected": false
           }
       ],
       [
           {
               "url": "http://pass.local/fcrepo/rest/repositories/4",
               "selected": true
           },
           {
               "url": "http://pass.local/fcrepo/rest/repositories/5",
               "selected": false
           }
       ]
   ],
   "optional": [
       {
           "url": "http://pass.local/fcrepo/rest/repositories/6",
           "selected": true
       }
   ]
}
```

In the above example:

* Repository 1 is required.
* At least one of repositories 2 and 3 are required, repository 2 should be the default choice.
* At least one of repositories 4 and 5 are required, repository 4 should be the default choice.
* Repository 6 is optional, but selected by default.

The json document contains the following fields:

* `required`: lists all repositories for which deposit is required
* `one-of`:  contains a list of lists of repositories.  Each top level list defines a group, from which one of the repositories listed within must be chosen.  Multiple groups may be returned, each containing independent lists of repositories.  These lists may intersect (i.e. the same repository can appear in multiple groups).
* `optional`:  lists all repositories for which deposit is completely optional

Repositories contained within each of the above lists are JSON objects containing the following fields:

* `url`: the URL to the repository resource in Fedora
* `selected`: optional field.  Specifies if the repository should be selected by default in the UI or not.
