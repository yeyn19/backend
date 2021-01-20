

# License.

This project is licensed under [GNU AFFERO GENERAL PUBLIC LICENSE Version 3](./license.md).

# Contribution

Our document is still under construction.

# Code style.
All code must be formatted by go fmt.

All tests of app/controllers is running with the same in-memory database,
so they shouldn't rely on a clean database and shouldn't cleanup after runs.
Also, they should be running under parallel mode.

All other tests should make it own database and clean it up after running.

# Permissions

Here are the permissions and their descriptions.

|         Name         |                                                 Description                                                 |
|:--------------------:|:-----------------------------------------------------------------------------------------------------------:|
|      read_user       |                                        the permission to read users                                         |
|     manage_user      |                                       the permission to manage users                                        |
|    create_problem    |                                               create problem                                                |
|    update_problem    | update problem. A scoped update_problem can only update selected problem. unscoped can update all problems. |
|    delete_problem    |                                      delete a problem. same as above.                                       |
| read_problem_secrets |                                read sensitive information such as test case.                                |
|      read_logs       |                                                 read logs.                                                  |
