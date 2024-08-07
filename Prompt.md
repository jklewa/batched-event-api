# Transformer Take Home Coding Exercise
Build a webservice in Golang that fulfills the documented API and Requirements listed below.

Provide a `Dockerfile` and `docker-compose.yml`.
It is assumed that your application will be run with this command: `docker-compose up`.

Please be prepared to present your solution, on your system, show it running,
and make minor modifications on request.


## Requirements

### Input
Please build a webservice that will receive a `POST` HTTP request.

The payload is User Event data, and will be in newline-delimited JSON, of which
examples are provided.

The size of the payload could be larger than the memory allocated to the webservice process.

The format and fields within the payload are not dynamic, and not expected to change any time soon.

It can be assumed that the rows are already ordered by their `time` field, both within a request and
across multiple requests as well. Meaning that the first row in a new request will have the same
time value or later, than the last row in the previous request.

For the sake of simplicity, it can also be assumed that new requests won't be made until the
previous request has received a response. In other words, only one request will be in-flight at a
time.

### Output
The webservice should respond with a `200` OK as soon as it has finished reading the request
payload.

The service should convert the received newline-delimited JSON to CSV, and write it to disk.

The files should be batched into 5 minute intervals, based on the `time` field in the rows of the
json request payload. The files should be named according to the first data point in them.

Multiple requests may contribute to the same files. For example, multiple requests with data that
fits into the same 5-minute interval, should result in only 1 CSV file written, even if there is a
long delay between requests. A request may also span multiple 5-minute intervals, resulting in
multiple CSV files written.

Once the files are closed, they are considered committed, and are no longer available for further
modification. In other words, you will need to keep the file open until new data coming in has
passed the 5-minute interval.


### Examples
Two `POST`s are made:
* The first payload contains user event rows timestamped starting at `2024-07-01T02:03:04Z`,
  with events occurring every second, with the last event at `2024-07-01T02:11:05`.
* The second payload contains user event rows timestamped starting at `2024-07-01T02:12:06`, 
  with events occurring every second, with the last event at `2024-07-01T02:15:07`.

The result should be 3 CSV files:
* The first CSV should have all data that is [`2024-07-01T02:03:04Z`, `2024-07-01T02:08:04Z`).
* The second CSV should have all data that is [`2024-07-01T02:08:04Z`, `2024-07-01T02:13:04Z`).
* The third CSV should have all data that is [`2024-07-01T02:13:04Z`, `2024-07-01T02:18:04Z`).


## FAQ

### How long should I spend on this exercise?
We have designed this exercise to take approximately 2 hours. Please complete what you can in
roughly that amount of time. We value a running solution that gets close or only meets some of the
requirements, over an unfinished but superior solution.

### How should I submit my solution?
A private repository would be ideal. We will work with you to provide access to the interviewers.

### Should I use any specific libraries?
This exercise was designed to be completed using only the Golang SDK.
You are free to use any open-source libraries you'd like though, as long as you can explain and
justify them.
