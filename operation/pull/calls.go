package pull

import "gitea.com/gitea/gitea-mcp/pkg/tool"

var pullReadMethods = []tool.MethodSpec{
	tool.Method("list_pull_requests"), tool.Method("get", "pull_number"),
	tool.Method("get_diff", "pull_number"), tool.Method("get_files", "pull_number"),
	tool.Method("get_status", "pull_number"), tool.Method("get_reviews", "pull_number"),
	tool.Method("get_review", "pull_number", "review_id"),
	tool.Method("get_review_comments", "pull_number"),
}

var pullWriteMethods = []tool.MethodSpec{
	tool.Method("create", "head", "base", "title"), tool.Method("update", "pull_number"),
	tool.Method("close", "pull_number"), tool.Method("reopen", "pull_number"),
	tool.Method("merge", "pull_number"), tool.Method("update_branch", "pull_number"),
	tool.Method("add_reviewers", "pull_number"), tool.Method("remove_reviewers", "pull_number"),
}
