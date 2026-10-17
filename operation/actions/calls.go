package actions

import "gitea.com/gitea/gitea-mcp/pkg/tool"

var runReadMethods = []tool.MethodSpec{
	tool.Method("list_workflows"), tool.Method("get_workflow", "workflow_id"),
	tool.Method("list_runs"), tool.Method("get_run", "run_id"),
	tool.Method("list_jobs"), tool.Method("list_run_jobs", "run_id"),
	tool.Method("get_job", "job_id"), tool.Method("get_job_log_preview", "job_id"),
	tool.Method("download_job_log", "job_id"), tool.Method("list_artifacts"),
	tool.Method("list_run_artifacts", "run_id"), tool.Method("get_artifact", "artifact_id"),
	tool.Method("download_artifact", "artifact_id"),
}

var runWriteMethods = []tool.MethodSpec{
	tool.Method("dispatch_workflow", "workflow_id", "ref"),
	tool.Method("cancel_run", "run_id"), tool.Method("rerun_run", "run_id"),
}
