package pkg

import (
	"reflect"
	"testing"
)

func TestReadBatchJob(t *testing.T) {
	type args struct {
		content []byte
	}
	tests := []struct {
		name string
		args args
		want JobList
	}{
		{
			name: "TestBasic",
			args: args{
				content: []byte(`BatchJobs:
  - Job: Example/Folder/mainbranch
    Variables:
      release-ticket: blah
      dry_run : false
      timeout : 600
      component: ops-atlantis
      version: v0.0.23
    Substitute:
      tenant: [tenant1, tenant2, tenant3, tenant4]`,
				),
			},
			want: JobList{
				Batch: []BatchJob{
					{
						Job: "Example/Folder/mainbranch",
						Variables: map[string]string{
							"release-ticket": "blah",
							"dry_run":        "false",
							"timeout":        "600",
							"component":      "ops-atlantis",
							"version":        "v0.0.23"},
						Substitute: map[string][]string{
							"tenant": {"tenant1", "tenant2", "tenant3", "tenant4"},
						},
					},
				},
			},
		},
		{
			name: "TestBasic",
			args: args{
				content: []byte(`
BatchJobs:
- Job: Example/Folder/mainbranch
  Substitute:
    component:
      - thing1
      - thing2
      - thing3
      - thing4
    tenant:
      - tenant1
      - tenant2
      - tenant3
      - tenant4
  Variables:
    component: ops-atlantis
    dry_run: false
    release-ticket: blah
    timeout: 600
    version: v0.0.23
      `,
				),
			},
			want: JobList{
				Batch: []BatchJob{
					{
						Job: "Example/Folder/mainbranch",
						Variables: map[string]string{
							"release-ticket": "blah",
							"dry_run":        "false",
							"timeout":        "600",
							"component":      "ops-atlantis",
							"version":        "v0.0.23"},
						Substitute: map[string][]string{
							"tenant":    {"tenant1", "tenant2", "tenant3", "tenant4"},
							"component": {"thing1", "thing2", "thing3", "thing4"},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReadBatchJob(tt.args.content); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadBatchJob() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateParameterList(t *testing.T) {
	type args struct {
		jobBatch BatchJob
	}
	tests := []struct {
		name string
		args args
		want []map[string]string
	}{{
		name: "TestBasic",
		args: args{
			jobBatch: BatchJob{
				Job: "Example/Folder/mainbranch",
				Variables: map[string]string{
					"release-ticket": "blah",
					"dry_run":        "false",
					"timeout":        "600",
					"component":      "ops-atlantis",
					"version":        "v0.0.23"},
				Substitute: map[string][]string{
					"tenant": {"tenant1", "tenant2"},
				}}},
		want: []map[string]string{
			{
				"tenant":         "tenant1",
				"release-ticket": "blah",
				"dry_run":        "false",
				"timeout":        "600",
				"component":      "ops-atlantis",
				"version":        "v0.0.23",
			},
			{
				"tenant":         "tenant2",
				"release-ticket": "blah",
				"dry_run":        "false",
				"timeout":        "600",
				"component":      "ops-atlantis",
				"version":        "v0.0.23",
			},
		},
	},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateParameterList(tt.args.jobBatch); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateParameterList() = %v, want %v", got, tt.want)
			}
		})
	}
}
