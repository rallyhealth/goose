# Goose

Wanna run jenkins in your cli? Use goose!


To install
```git clone git@github.com:rallyhealth/goose.git
 cd Goose
 go install .
```

## Command Summary


### run

Goose run will invoke a jenkins job based on the current git directory and branch.
If you want to pick a job manually, specify the path of the job.

```
goose run # Will list available jobs based on your directory

goose run Example/Folder/CreateThing/mainbranch # Run the CreateThing job on the mainbranch branch.
```

### jobs

Print out the list of jobs in a jenkins directory. (Root directory by default)

```
❯ goose jobs
Example
Example2
...

❯ goose jobs Example
SubJob1
SubJob2


```

### Latest

Latest will print out the output of the latest running job of the given job OR based on the current git/branch.


### Search

Search will list out the jenkins jobs relevant to your current git directory

### RunList

Runlist will run jobs concurrently from a batchfile. For example:

```
cat <<EOF >>myBatchFile
BatchJobs:
  - Job: Path/To/My/Job
    Variables:
      VariableA: True
      VariableB: False
    Substitute:
      VariableC: [value1,value2]
EOF

goose runlist myBatchFile
```

Will run 2 jobs concurrently for the `Path/To/My/Job` on the main branch. The `VariableA` and `VariableB` variables will be the same for these 4 job runs. The end result is that each job is run with the same variables with `VariableC` being different for each run.

## Requirements

You'll need to define JENKINS_EMAIL and JENKINS_API_KEY and JENKINS_ROOT and JENKINS_LOGIN_URL.
The api key needs to be made at JENKINS_ROOT/me/configure.

If you're using an enterprise jenkins setup, double check that you can reach all jobs from the root URL.
Your JENKINS_ROOT should have no path (i.e, https://ci.mycompany.com). Your JENKINS_LOGIN_URL contains the path to the root job directory (i.e, https://ci.mycompany.com/path/job/Base).

If you are using an enterprise Jenkins setup, the login url is most likely similar to https://ci.mycompany.com/cjob/job/BaseFolderName
