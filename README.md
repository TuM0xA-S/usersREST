# notes
simple rest api using echo framework and json file based database

### features
* modular organisation (db fully separated)
* db optimized for perfomance 
* good test coverage
* handles signals

### api reference
action              | request
------------------- | ---------------
create user		    | `POST /users` + json payload (1)
get user	    	| `GET /users/:id`
update user  	    | `PUT /users/:id` + json payload (1)
delete user  	    | `DELETE /users/:id`
get list of users   | `GET /users`

### explanations
1. json payload example `{"name": "John", "age": 20}`
2. response format `{"message": "<text>", "data": <User|[]User>}`

### usage
just run it(defaults will be used), also it has configuration with cmd args (-h for help)
