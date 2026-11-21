/**
 * @name Admin API route without admin middleware
 * @description An admin API registration path must inherit RequireAdmin from its own chi router.
 * @kind problem
 * @problem.severity error
 * @security-severity 8.1
 * @precision high
 * @id go/relay-admin-router-middleware
 * @tags security
 *       external/cwe/cwe-862
 */

import go

private string middlewarePackage() {
  result = "github.com/ThreeHats/foundryvtt-rest-api-relay/go-relay/internal/middleware"
}

private Expr receiver(CallExpr call) { result = call.getCalleeExpr().(SelectorExpr).getBase() }

private predicate chiCall(CallExpr call, string method) {
  call.getCalleeName() = method and
  call.getTarget().(Method).getReceiverBaseType().getPackage().getPath() =
    "github.com/go-chi/chi/v5"
}

// Trace local values to their constructor or callback parameter. Receiver calls
// create post-update nodes, so global value numbering alone cannot identify a
// chi router across successive registrations.
private predicate routerOrigin(DataFlow::Node source) {
  source.asExpr() instanceof CallExpr or source instanceof DataFlow::ParameterNode
}

private predicate sameRouter(Expr a, Expr b) {
  exists(DataFlow::Node source |
    routerOrigin(source) and
    DataFlow::localFlow(source, DataFlow::exprNode(a)) and
    DataFlow::localFlow(source, DataFlow::exprNode(b))
  )
}

// For a guard, possible aliasing is insufficient: a receiver reassigned on one
// branch must not authorize another router that shares only one possible origin.
private predicate definitelySameRouter(Expr a, Expr b) {
  sameRouter(a, b) and
  not exists(DataFlow::Node one, DataFlow::Node two |
    routerOrigin(one) and
    routerOrigin(two) and
    one != two and
    DataFlow::localFlow(one, DataFlow::exprNode(a)) and
    DataFlow::localFlow(two, DataFlow::exprNode(a))
  ) and
  forall(DataFlow::Node origin |
    routerOrigin(origin) and DataFlow::localFlow(origin, DataFlow::exprNode(b))
  |
    DataFlow::localFlow(origin, DataFlow::exprNode(a))
  )
}

private predicate requireAdmin(Expr e) {
  e.(CallExpr).getTarget().hasQualifiedName(middlewarePackage(), "RequireAdmin")
}

// A top-level Use before registration is unconditional on that router. A Use in
// a conditional block, a sibling callback, or after registration is not evidence.
private predicate guarded(CallExpr call) {
  chiCall(call, "With") and requireAdmin(call.getAnArgument())
  or
  exists(CallExpr use, ExprStmt stmt, FuncDef f |
    chiCall(use, "Use") and
    requireAdmin(use.getAnArgument()) and
    definitelySameRouter(receiver(use), receiver(call)) and
    f = call.getEnclosingFunction() and
    use.getEnclosingFunction() = f and
    stmt.getExpr() = use and
    stmt.getParent() = f.getBody() and
    (
      use.getLocation().getEndLine() < call.getLocation().getStartLine()
      or
      use.getLocation().getEndLine() = call.getLocation().getStartLine() and
      use.getLocation().getEndColumn() < call.getLocation().getStartColumn()
    )
  )
}

private predicate returnedRouter(FuncDecl f, Expr router) {
  exists(ReturnStmt ret | ret.getEnclosingFunction() = f and router = ret.getExpr(0))
}

// The child belongs to this registration path only: Group/Route callback,
// With result, or the router returned by a statically resolved mounted factory.
private predicate childRouter(CallExpr parent, Expr child) {
  exists(FuncLit callback |
    (chiCall(parent, "Group") or chiCall(parent, "Route")) and
    callback = parent.getAnArgument() and
    child = callback.getParameter(0).getAReference()
  )
  or
  chiCall(parent, "With") and child = parent
  or
  exists(CallExpr factory |
    chiCall(parent, "Mount") and
    DataFlow::localFlow(DataFlow::exprNode(factory), DataFlow::exprNode(parent.getArgument(1))) and
    returnedRouter(factory.getTarget().getFuncDecl(), child)
  )
}

private string handlerPackage() {
  result = "github.com/ThreeHats/foundryvtt-rest-api-relay/go-relay/internal/handler"
}

private predicate adminRouter(FuncDecl admin) {
  admin.getFunction().hasQualifiedName(handlerPackage(), "AdminRouter")
}

private string routePath(CallExpr call) {
  if chiCall(call, ["Method", "MethodFunc"])
  then result = call.getArgument(1).(StringLit).getValue()
  else result = call.getArgument(0).(StringLit).getValue()
}

bindingset[path]
private predicate apiPath(string path) { path = "/api" or path.matches("/api/%") }

// Scope is a finite state, avoiding unbounded concatenation through recursive
// mounted factories. Only the admin root can introduce the /api boundary.
private string childScope(CallExpr parent, string scope) {
  scope = ["root", "api", "deep-api", "other"] and
  if chiCall(parent, ["Route", "Mount"])
  then
    exists(string path |
      path = routePath(parent) and
      (
        if scope = "root"
        then
          if path = "/api"
          then result = "api"
          else
            if apiPath(path)
            then result = "deep-api"
            else result = "other"
        else
          if scope = ["api", "deep-api"]
          then result = "deep-api"
          else result = "other"
      )
    )
  else result = scope
}

// Keep API scope and protection on the same path. A factory mounted on both a
// protected and an unprotected branch must still report its unprotected path.
private predicate unprotected(CallExpr call, string scope) {
  not guarded(call) and
  (
    exists(FuncDecl admin, Expr root |
      adminRouter(admin) and
      returnedRouter(admin, root) and
      sameRouter(root, receiver(call)) and
      scope = "root"
    )
    or
    exists(CallExpr parent, Expr child, string parentScope |
      unprotected(parent, parentScope) and
      childRouter(parent, child) and
      sameRouter(child, receiver(call)) and
      scope = childScope(parent, parentScope)
    )
  )
}

private predicate reportException(CallExpr call, string scope) {
  chiCall(call, "Get") and
  call.getArgument(1)
      .(FunctionName)
      .getTarget()
      .hasQualifiedName(handlerPackage(), "TestReportHandler") and
  (
    scope = "api" and call.getArgument(0).(StringLit).getValue() = "/tests/report"
    or
    scope = "root" and call.getArgument(0).(StringLit).getValue() = "/api/tests/report"
  )
}

from CallExpr registration, string scope
where
  chiCall(registration,
    [
      "Get", "Post", "Put", "Patch", "Delete", "Head", "Options", "Connect", "Trace", "Handle",
      "HandleFunc", "Method", "MethodFunc"
    ]) and
  unprotected(registration, scope) and
  (
    scope = ["api", "deep-api"]
    or
    scope = "root" and apiPath(routePath(registration))
  ) and
  not reportException(registration, scope)
select registration, "This admin API route has a registration path without RequireAdmin middleware."
