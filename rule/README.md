# Policy rules DSL

A DSL is proposed which (a) determines which policies are applicable to a submission, (b) determines which repositories are relevant to the policies, and (c) contains information which the policy service can use to sort repositories into.

This DSL contains a `$schema` field that specifies the schema of the DSL being used (it almost certainly will evolve), and an `include-policies` field which has a list of rules for including policies.  The DSL file is a configuration file provided to the policy service.

An example of JHU's current policy rules is as follows:

```json
{
    "$schema": "https://oa-pass.github.io/pass-policy-service/schemas/policy_config_1.0.json",
    "policy-rules": [
        {
            "description": "Must deposit to one of the repositories indicated by primary funder",
            "policy-id": "${submission.grants.primaryFunder.policy}",
            "type": "funder",
            "repositories": [
                {
                    "repository-id": "${policy.repositories}"
                }
            ]
        },
        {
            "description": "Must deposit to one of the repositories indicated by direct funder",
            "policy-id": "${submission.grants.directFunder.policy}",
            "type": "funder",
            "repositories": [
                {
                    "repository-id": "${policy.repositories}"
                }
            ]
        },
        {
            "description": "Members of the JHU community must deposit into JScholarship, or some other repository.",
            "policy-id": "/policies/5e/2e/16/92/5e2e1692-c128-4fb4-b1a0-95c0e355defd",
            "type": "institution",
            "conditions": [
                {
                    "endsWith": {
                        "@johnshopkins.edu": "${header.Ajp_eppn}"
                    }
                }
            ],
            "repositories": [
                {
                    "repository-id": "/repositories/41/96/0a/92/41960a92-d3f8-4616-86a6-9e9cadc1a269",
                    "selected": true
                },
                {
                    "repository-id": "*"
                }
            ]
        }
    ]
}
```

In this example:

* All policies for funders of grants associated with the submission (direct or primary) will be included
  * As far as repositories are concerned, all repositories listed under of a role will be grouped into a "one-of" bucket.   For example, if `${policy.repositories.id}` expands to two repositories (by way of a policy linking to two repositories), then both repositories are in the same one-of bucket.  The policy service will be be smart enough to promote one-of groups containing only a single item to the required list, or demoting repositories in a one-of bucker to optional if all any other repository in the one-of bucket is required by a different policy. The jScholarship policy will be added only for users for whom the request `Eppn` header contains the substring `@johnshopkins.edu`.  Effectively, this means everybody from JHU will get this policy added.
  * As far as its repository is concerned, it is required by default, but becomes optional if any other repository is required.  

The top-level fields in the DSL are:

* `$schema`:  Required.  It must point to a known JSON schema for the policy service DSL
* `policy-rules`:  Contains a list of policy inclusion rules

Policy inclusion rules are JSON objects containing the following fields:

* `description`:  A human readable description of the rule.  Optional.
* `policy-id`:  a string containing a a single policy URI, or a variable substitution resulting in one or more policy URIs
  * In the case of a variable substitution resulting in many URIs, it is equivalent to creating multiple policy rules, each one containing a single policy-id from that list.  
repositories:  contains a list of repository description JSON objects, specifying which repositories satisfy the given policy.
* `condition`:  Optional.  JSON object describing a condition where the policy is included only if the condition evaluates to true.  If this field is not present, it is presumed that inclusion of the policy is unconditional.  See the schema for more details, but conditions include:
  * `equals`: true if two strings are equal
  * `endsWith`: true if a string ends with another
  * `contains`: true if a string contains another as a substring
  * `anyOf`: true if any of the given list of conditions are true
  * `noneOf`: true if none of the given list of conditions are true

Repositories are JSON objects with the following fields:

* `repository-id`: the URI of the repository resource in Fedora, or `*` to mean "any".
* `selected`:  (optional boolean) if true, the repository will be indicated as "selected" by default in the result to Ember.

## Variable substitution

Any key or value of the form `${variable}` is a variable.  

At time of rule evaluation, the following variables are available:

* `submission`:  the submission object
* `header`:  the list of Http headers in the request, including all shibboleth headers.

Variables can use dot notation, which means different things in context

* For headers, header.NAME means the value of the header NAME
* For repository objects, object.FIELD means the value of the field FIELD
  * If FIELD is a uri pointing to another repository object, then `${object.FIELD}` is itself a repository object
  * If FIELD is a string, then `${object.FIELD}` is a string.  Further dots will attempt to treat that string as a JSON blob (e.g. ${submission.metadata.author})
  * if FIELD is a list of URIs  of objects, then `${object.FIELD}` is a list of objects.

A graph of objects can be navigated via dot notation, e.g. `${submission.grants.primaryFunder}` is a list of Funder objects.  `${submission.grants.primaryFunder.id}` is a list of URIs.

In the case of policy rules, where a `policy-id` is a list of URIs, the variables available to repositories block are based of one matching value.  You can imagine this translating to N rules (each one with a policy URI from the list), with each repository block inheriting the values of the policy rule that contains it.  For example:

```json
 {
        "description": "Must deposit to one of the repositories indicated by primary funder",
        "policy-id": "${submission.grants.primaryFunder.policy}",
        "repositories": [
            {
                "repository-id": "${policy.repositories}",
            }
        ]
}
```

In this case for each matching policy, the value of `${submission.grants.primaryFunder.policy}` is fixed inside the repositories block.  That is to say, submission is the given submission object, (as it always is), `${submission.grants}` is the single grant object used in producing the `${submission.grants.primaryFunder.policy}` value for this particular policy, etc.

As a shortcut, `${policy}` is an alias for `${submission.grants.primaryFunder.policy}`, and is a repository object.  Any such dot segment can function as an alias, as long as it is unambiguous
