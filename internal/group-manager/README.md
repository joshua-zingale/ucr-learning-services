A group management web server that is intended to serve as middleware.

# Plan

## Behavior

The group manager should receive a web request with an X-Email header containing a unique email for the user.
The manager will then determine to which groups the email belongs,
responding with a copy of the request,
only with an X-Groups header added that contains a comma separated list of permissions like

```
X-Groups: cs100.instructor,cs287.student
```

To minimize the size of the header, a subset of all roles can be obtained.
If the root URL is `http://127.0.0.1:6400/`, then proxying `http://127.0.0.1:6400/cs100`
would add to the headers, for the same hypothetical user as above,

```
X-Groups: instructor
```


## Configuration


### Initialization 

The policy manager web server should support initialization via a YAML file like

```
cs100:
    instructor:
        - bob@ucr.edu
    student:
        - tom@ucr.edu
        - dick@ucr.edu
        - harry@ucr.edu
cs9B: ...
```

This will assign bob@ucr.edu, for example, the group cs100.instructor


### Updates

All methods for updating the group assignments must function while keeping the group manager functional.


#### Command Line Interface

The group manager should have a terminal interface that supports

```bash
group-manager create -p cs100.instructor # create the parent groups
group-manager create cs100.students # fail if the parent(s) are not already created
group-manager destroy cs100.students # delete the students group in cs100 if it is empty
group-manager destroy -r cs100 # require -r for recursive deletes
group-manager assign cs100.instructor bob@ucr.edu rob@ucr.edu # Add one or more emails to a group
group-manager unassign cs100.instructor bob@ucr.edu rob@ucr.edu # remove one or more emails from a group
group-manager get-groups bob@ucr.edu rob@ucr.edu # get the intersection of the goups held by all specified emails
group-manager list cs100 # get all subgroups, e.g. "instructor, TA, student"
```

#### YAML File

The system should be settable to track a YAML file, causing the group assignments to reload whenever the tracked YAML file has been updated.


