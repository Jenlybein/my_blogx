import json
import pathlib


ROOT = pathlib.Path(r"e:\project\blogx\blogx_server")
OPENAPI = ROOT / ".read" / "myblogx.openapi.json"


def has_bearer(op: dict) -> bool:
    return any("bearerAuth" in item for item in (op.get("security") or []))


def has_refresh_cookie(op: dict) -> bool:
    return any("refreshTokenCookie" in item for item in (op.get("security") or []))


def auth_suffix(op: dict) -> str:
    if has_refresh_cookie(op):
        return "调用时需要携带 HttpOnly refresh_token Cookie。"
    if has_bearer(op):
        return "调用时需要在请求头中携带 Bearer access token。"
    return ""


OP_OVERRIDES = {
    (
        "POST",
        "/api/users/login",
    ): "使用用户名或邮箱加密码登录。成功后返回 access token，后续受保护接口应通过 Authorization: Bearer <token> 调用。\n\n调用顺序：\n1. 提交用户名或邮箱与密码。\n2. 后端校验账号、密码与风控状态。\n3. 成功后返回 access token，并写入 refresh_token Cookie。\n4. 后续受保护接口优先使用 Bearer token 调用，access token 过期后再调用刷新接口。",
    (
        "POST",
        "/api/users/refresh",
    ): "通过 HttpOnly refresh_token Cookie 换取新的 access token。当前接口主要用于 access token 过期后的续期，不需要在 body 中再传 token。\n\n调用顺序：\n1. 前端检测 access token 失效或即将失效。\n2. 携带浏览器自动附带的 refresh_token Cookie 调用本接口。\n3. 成功后拿到新的 access token。\n4. 用新的 access token 重试原先的受保护请求。",
    ("POST", "/api/users/logout"): "退出当前设备登录态。通常在前端用户主动退出时调用，会清理当前会话对应的登录状态。",
    ("POST", "/api/users/logout/all"): "退出当前账号在所有设备上的登录态。适合修改密码后或用户希望强制全端下线时调用。",
    (
        "POST",
        "/api/users/email/verify",
    ): "发送邮箱验证码。不同 type 表示不同业务场景，调用成功后会返回 email_id，后续校验验证码时需要携带该值。\n\n调用顺序：\n1. 前端提交邮箱和验证码业务类型。\n2. 后端发送邮件并返回 email_id。\n3. 前端提示用户输入验证码。\n4. 后续在注册、登录、绑定邮箱或找回密码接口中提交 email_id 与 email_code。",
    ("POST", "/api/users/email/register"): "使用邮箱验证码完成注册。应先调用邮箱验证码发送接口拿到 email_id，再提交验证码与密码。",
    (
        "POST",
        "/api/users/email/login",
    ): "先调用邮箱验证码发送接口，再使用 email_id 与 email_code 完成登录。成功后返回 access token。\n\n调用顺序：\n1. 先调用邮箱验证码发送接口拿到 email_id。\n2. 用户输入邮件中的验证码。\n3. 调用本接口提交 email_id 与 email_code。\n4. 成功后拿到 access token，并按常规登录态使用。",
    ("POST", "/api/users/qq"): "使用 QQ 登录授权 code 换取系统登录态。成功后返回 access token，并写入 refresh_token Cookie。",
    ("GET", "/api/users/detail"): "获取当前登录用户自己的完整资料信息，适合个人中心初始化时调用。",
    ("GET", "/api/users/base"): "根据用户 ID 获取用户页公开基础信息，用于个人主页、作者卡片等展示场景。",
    ("GET", "/api/users/login/log"): "获取指定用户的登录记录列表，可按用户或记录类型筛选。",
    ("PUT", "/api/users/password/renewal/email"): "已登录用户通过邮箱验证码修改密码。成功后会使旧登录态失效。",
    ("PUT", "/api/users/password/recovery/email"): "未登录或忘记密码场景下，通过邮箱验证码重置密码。成功后旧登录态会失效。",
    ("PUT", "/api/users/email/bind"): "为当前登录用户绑定邮箱地址。应先获取邮箱验证码，再提交 email_id 与验证码。",
    ("PUT", "/api/users/info"): "更新当前登录用户资料，例如昵称、头像、简介和个人展示配置。",
    ("GET", "/api/site/qq_url"): "获取 QQ 登录跳转地址。前端在发起 QQ 授权登录前应先调用该接口。",
    ("GET", "/api/site/ai_info"): "获取站点 AI 助手公开配置，例如昵称、头像、是否启用等，用于前端初始化展示。",
    ("GET", "/api/articles"): "获取文章列表。可结合 type、status、user_id、分页与排序参数筛选结果。",
    (
        "POST",
        "/api/articles",
    ): "发布新文章。请求体中的 content 为正文内容，tag_ids 为已存在标签 ID 列表，cover 为封面图片地址。\n\n调用顺序：\n1. 前端先完成图片上传，拿到可用图片 URL。\n2. 组织标题、正文、标签、分类、封面和状态等字段。\n3. 调用本接口提交文章。\n4. 成功后再跳转到文章详情页或管理页。",
    ("PUT", "/api/articles/{id}"): "更新指定文章。通常用于文章编辑页保存，未提交的字段会按接口实现使用新值覆盖旧值。",
    ("GET", "/api/articles/{id}"): "获取指定文章详情。适合文章详情页初始化，也会返回文章统计与作者信息。",
    ("DELETE", "/api/articles/{id}"): "删除当前用户自己的文章。删除后文章不再对外展示。",
    ("POST", "/api/articles/view"): "记录文章浏览量与访问历史。文章详情页打开后通常调用一次。",
    ("GET", "/api/articles/history"): "获取当前用户的文章访问历史列表。",
    ("DELETE", "/api/articles/history"): "删除当前用户的文章访问历史记录，可按请求体中的 ID 列表批量删除。",
    ("POST", "/api/articles/favorite"): "将文章加入指定收藏夹，或在收藏夹之间建立关联。",
    ("PUT", "/api/articles/favorite"): "更新收藏夹基本信息，如标题、封面和简介。",
    ("DELETE", "/api/articles/favorite"): "删除收藏夹。删除前请确认前端不再引用该收藏夹。",
    ("GET", "/api/articles/favorite"): "获取收藏夹列表，可根据 type 与 user_id 查询自己的收藏夹或公开收藏夹。",
    ("GET", "/api/articles/favorite/contents"): "获取某个收藏夹中的文章列表。需要提供 favorite_id。",
    ("DELETE", "/api/articles/favorite/contents"): "批量移除收藏夹中的文章关联关系，不会删除文章本身。",
    ("POST", "/api/articles/category"): "创建或更新文章分类，具体行为取决于请求体中的 ID 与字段。",
    ("DELETE", "/api/articles/category"): "删除分类。删除前请确认没有业务继续依赖该分类。",
    ("GET", "/api/articles/category"): "获取分类列表，可结合 type 等查询参数控制返回范围。",
    ("GET", "/api/articles/category/options"): "获取分类选项数据，适合下拉框或表单选择器使用。",
    ("GET", "/api/articles/tags"): "获取标签列表，可用于后台管理、筛选和页面展示。",
    ("GET", "/api/articles/tags/options"): "获取标签选项数据，适合表单选择组件初始化时调用。",
    ("GET", "/api/articles/top"): "获取置顶列表，可按用户或类型筛选。",
    ("POST", "/api/articles/top"): "设置文章置顶关系，用于用户置顶或管理员置顶。",
    ("DELETE", "/api/articles/top"): "移除文章置顶关系。",
    ("POST", "/api/comments"): "创建评论或回复评论。请求体中的父评论信息决定是根评论还是二级回复。",
    ("GET", "/api/comments"): "获取文章根评论列表，需要提供 article_id。",
    ("GET", "/api/comments/replies"): "获取某条根评论下的回复列表，需要提供 article_id 和 root_id。",
    ("GET", "/api/comments/man"): "获取后台管理评论列表，支持按文章、用户、状态和类型筛选。",
    ("DELETE", "/api/comments/{id}"): "删除指定评论。支持用户删除自己的评论或管理员执行管理删除。",
    ("POST", "/api/comments/{id}/digg"): "对指定评论执行点赞或取消点赞操作。",
    ("GET", "/api/sitemsg/conf"): "获取当前用户的站内消息配置。",
    ("PUT", "/api/sitemsg/conf"): "更新当前用户的站内消息配置，例如是否接收特定类型通知。",
    ("GET", "/api/sitemsg"): "获取站内消息列表，可结合分页和类型参数筛选。",
    ("POST", "/api/sitemsg"): "将指定站内消息标记为已读。",
    ("DELETE", "/api/sitemsg"): "删除站内消息，可按请求体中的消息 ID 列表批量删除。",
    ("GET", "/api/sitemsg/user"): "获取当前用户未读消息统计，用于顶部角标或消息中心入口提示。",
    ("GET", "/api/global_notif"): "获取全局通知列表，可按类型与分页参数筛选。",
    ("DELETE", "/api/global_notif/user"): "用户侧删除自己已接收的全局通知记录。",
    ("POST", "/api/global_notif/read"): "将指定全局通知标记为已读。",
    ("GET", "/api/follow"): "获取关注列表，可查看指定用户关注了哪些人。",
    ("GET", "/api/fans"): "获取粉丝列表，可查看哪些用户关注了目标用户。",
    ("POST", "/api/follow/{id}"): "关注指定用户。路径中的 id 为被关注用户 ID。",
    ("DELETE", "/api/follow/{id}"): "取消关注指定用户。路径中的 id 为被取消关注用户 ID。",
    ("GET", "/api/chat/sessions"): "获取私信会话列表，可按 type 区分不同会话视图。",
    ("DELETE", "/api/chat/sessions"): "删除私信会话，通常用于会话列表页的删除操作。",
    ("GET", "/api/chat/messages"): "获取私信消息列表，需要提供 session_id 或用户相关筛选条件。",
    ("DELETE", "/api/chat/messages"): "删除私信消息。当前接口通常按请求体提供的消息范围执行删除。",
    ("POST", "/api/chat/read"): "将指定会话或消息标记为已读。",
    ("GET", "/api/chat/ws"): "使用 ticket 建立聊天 WebSocket 连接。ticket 需先通过已登录接口换取。",
    ("GET", "/api/search/articles"): "按关键字、标签和排序方式搜索文章列表。适合搜索页或站内搜索框使用。",
    ("POST", "/api/ai/metainfo"): "基于文章正文生成标题、摘要、标签等元信息，适合文章编辑器辅助生成。",
    ("POST", "/api/ai/search/list"): "使用自然语言对站内文章进行 AI 搜索，并返回结构化列表结果。",
    ("POST", "/api/ai/search/llm"): "站点 AI 助手对话接口。当前为流式或增量式场景时，前端需要按接口返回约定处理。",
    ("GET", "/api/data/sum"): "获取后台汇总统计数据，例如文章数、用户数、评论数等。",
    ("GET", "/api/data/growth"): "获取增长趋势数据，可按 type 指定统计口径。",
    ("GET", "/api/data/article-year"): "获取按年份聚合的文章统计数据。",
    ("GET", "/api/imagecaptcha"): "获取图片验证码，用于登录、注册或其他需要图形校验的场景。",
    ("GET", "/api/images"): "获取图片资源列表，支持按分页参数分页查询。",
    ("DELETE", "/api/images"): "删除图片资源。接口会同步删除对象存储文件、图片记录和图片引用记录。",
    (
        "POST",
        "/api/images/upload-tasks",
    ): "前端先计算 qetag/hash，再调用本接口进行预去重并获取七牛上传票据。若返回 skip_upload=true，则可直接使用返回的 url。\n\n调用顺序：\n1. 前端选中图片后先计算 qetag/hash。\n2. 调用本接口提交文件名、大小、mime 和 hash。\n3. 若命中预去重，直接使用返回的 url，不再上传。\n4. 若未命中，则拿到 upload_token、object_key 等信息上传到七牛。\n5. 正式环境上传后依赖七牛 callback，再轮询上传状态接口拿结果；开发环境可直接调用 complete 接口兜底确认。",
    (
        "GET",
        "/api/images/upload-tasks/{id}",
    ): "查询上传任务状态。正式环境主要配合七牛 callback 轮询最终结果，状态变为 ready 后即可使用返回的 url。\n\n调用顺序：\n1. 前端调用创建上传任务接口并完成文件直传。\n2. 正式环境等待七牛回调触发后端确认。\n3. 前端周期性轮询本接口。\n4. 当 status 变为 ready 时，读取 url 作为最终图片地址；若为 failed，则根据 error_msg 提示用户。",
    (
        "POST",
        "/api/images/upload-tasks/complete",
    ): "手动完成图片上传任务。主要用于开发调试，或七牛回调不可达时的兜底确认。\n\n调用顺序：\n1. 前端创建上传任务并完成七牛直传。\n2. 如果当前环境没有走 callback，可调用本接口。\n3. 后端会校验对象、确认图片信息并写入正式图片记录。\n4. 成功后直接从响应里读取图片 url。",
    (
        "POST",
        "/api/images/qiniu/callback",
    ): "七牛上传完成后回调本接口，服务端据此完成图片最终确认。通常正式环境主要依赖这条链路。\n\n调用顺序：\n1. 前端先调用创建上传任务接口并完成七牛直传。\n2. 七牛上传成功后回调本接口。\n3. 后端完成对象校验、图片确认和入库。\n4. 前端再通过上传状态查询接口轮询 ready 结果。",
    (
        "POST",
        "/api/images/qiniu/audit/callback",
    ): "七牛内容审核完成后回调本接口，服务端会根据审核建议更新图片状态。\n\n调用顺序：\n1. 图片上传并确认完成后，七牛异步进行内容审核。\n2. 审核结果通过本接口回调到后端。\n3. 后端把图片状态更新为 ready、reviewing 或 blocked。\n4. 后续业务可根据图片状态决定是否允许继续展示或引用。",
    ("GET", "/api/site/{name}"): "获取指定站点公开配置。路径中的 name 用于区分 site、qq、ai 等公开配置项。",
    ("PUT", "/api/site/{name}"): "按路径中的 name 更新对应站点配置。不同 name 对应不同 JSON 结构，请按字段说明传值。",
    ("GET", "/api/site/admin/{name}"): "管理员获取后台站点配置详情。路径中的 name 用于区分 site、email、qq、qiniu、ai 等配置。",
}


def generic_desc(method: str, summary: str, op: dict) -> str:
    auth = auth_suffix(op)
    if method == "GET":
        if "列表" in summary:
            base = f"{summary}。可结合 query 参数进行筛选，并在提供分页参数时按页返回结果。"
        elif "详情" in summary:
            base = f"{summary}。按 path 或 query 指定目标资源后返回详细信息。"
        elif "选项" in summary:
            base = f"{summary}。适合表单下拉框、筛选器或初始化选项数据时调用。"
        else:
            base = f"{summary}。请按接口定义提供 path 或 query 参数。"
    elif method == "POST":
        if "登录" in summary:
            base = f"{summary}。提交 JSON 请求体完成认证，成功后请按响应数据继续后续流程。"
        elif "创建" in summary or "发布" in summary:
            base = f"{summary}。请按 JSON 请求体提供业务字段，成功后会返回创建结果。"
        elif "设置" in summary or "标记" in summary:
            base = f"{summary}。该接口会对目标资源执行状态变更。"
        else:
            base = f"{summary}。请按 JSON 请求体提供所需参数。"
    elif method == "PUT":
        base = f"{summary}。通常用于更新已有资源，请按 JSON 请求体提供需要修改的字段。"
    elif method == "DELETE":
        base = f"{summary}。请确认目标资源无误后再调用，删除后通常不可恢复。"
    else:
        base = f"{summary}。"
    return (base + (" " + auth if auth else "")).strip()


PARAM_DESC = {
    "id": "目标资源的主键 ID，用于定位具体记录。",
    "name": "路径中的配置名，用于选择具体站点配置项。",
    "page": "分页页码，从 1 开始。",
    "limit": "每页返回条数。",
    "type": "类型筛选参数，具体取值请结合该接口业务含义使用。",
    "status": "状态筛选参数，具体取值请结合该接口业务含义使用。",
    "user_id": "目标用户 ID，用于筛选或指定用户。",
    "article_id": "目标文章 ID。",
    "favorite_id": "目标收藏夹 ID。",
    "root_id": "根评论 ID，用于获取该根评论下的回复。",
    "session_id": "会话 ID，用于定位私信会话。",
    "order": "排序方式参数，具体格式按接口实现约定。",
    "sort": "搜索结果排序方式。",
    "key": "搜索关键字或查询词。",
    "tag_list": "参与筛选的标签名称列表。",
    "t": "消息筛选类型参数。",
    "title": "标题文本或标题筛选值。",
    "ticket": "WebSocket 建连票据，需先通过已登录接口换取。",
    "log_type": "日志类型筛选值。",
    "level": "日志级别筛选值。",
    "ip": "IP 地址筛选值。",
    "login_status": "登录结果状态筛选值。",
    "service_name": "服务名筛选值。",
}


COMMON_FIELD_DESC = {
    "code": "业务状态码，0 通常表示成功，其它值表示不同失败场景。",
    "msg": "接口返回消息，通常用于提示当前请求结果。",
    "id": "主键 ID。",
    "id_list": "需要批量操作的主键 ID 列表。",
    "created_at": "记录创建时间。",
    "updated_at": "记录最后更新时间。",
    "deleted_at": "软删除时间；为空表示未删除。",
    "user_id": "关联的用户 ID。",
    "article_id": "关联的文章 ID。",
    "favorite_id": "关联的收藏夹 ID。",
    "session_id": "关联的会话 ID。",
    "root_id": "关联的根评论 ID。",
    "cover": "封面图片地址。",
    "avatar": "头像图片地址。",
    "title": "标题文本。",
    "abstract": "摘要或简介文本。",
    "content": "正文内容，通常为 Markdown 或富文本内容。",
    "username": "用户名。",
    "nickname": "昵称。",
    "password": "密码明文输入字段，仅用于请求提交。",
    "pwd": "密码明文输入字段，仅用于请求提交。",
    "old_password": "旧密码。",
    "new_password": "新密码。",
    "email": "邮箱地址。",
    "email_id": "邮箱验证码会话 ID，用于和验证码配对校验。",
    "email_code": "邮箱验证码。",
    "ip": "IP 地址。",
    "addr": "IP 归属地或解析后的地址信息。",
    "ua": "用户代理字符串。",
    "user_nickname": "关联用户昵称。",
    "user_avatar": "关联用户头像地址。",
    "author_name": "作者昵称。",
    "author_username": "作者用户名。",
    "author_avatar": "作者头像地址。",
    "category_id": "分类 ID。",
    "category_name": "分类名称。",
    "category_title": "分类标题。",
    "tag_ids": "标签 ID 列表，需传入已存在标签的主键。",
    "tags": "标签列表。",
    "view_count": "浏览次数。",
    "digg_count": "点赞次数。",
    "comment_count": "评论数。",
    "favor_count": "收藏数。",
    "comments_toggle": "是否开启评论。",
    "status": "当前资源状态值。",
    "type": "当前对象的类型值。",
    "show": "是否展示该资源。",
    "href": "跳转链接地址。",
    "enable": "是否启用。",
    "bucket": "对象存储桶名称。",
    "provider": "对象存储提供商标识。",
    "object_key": "对象存储中的对象键。",
    "file_name": "原始文件名。",
    "mime_type": "文件 MIME 类型。",
    "size": "文件大小，单位通常为字节。",
    "hash": "文件内容标识，当前图片上传链路中使用 qetag/hash。",
    "upload_id": "上传任务 ID。",
    "upload_token": "上传到对象存储所需的临时票据。",
    "region": "对象存储区域标识。",
    "expire_at": "上传票据过期时间。",
    "max_size": "允许上传的最大文件大小。",
    "image_id": "图片资源 ID。",
    "url": "资源可访问地址。",
    "error_msg": "失败原因说明。",
    "skip_upload": "为 true 时表示命中预去重并跳过实际上传。",
    "role": "用户角色值。",
    "place": "地区或所在地信息。",
    "code_age": "账号注册时长或站龄信息。",
    "like_tags": "用户偏好标签列表，用于个性化展示。",
    "favorites_visibility": "收藏夹可见性配置。",
    "followers_visibility": "关注列表可见性配置。",
    "fans_visibility": "粉丝列表可见性配置。",
    "home_style_id": "个人主页样式 ID。",
    "register_source": "注册来源。",
    "updated_username_date": "最近一次修改用户名的时间。",
    "log_type": "日志类型。",
    "level": "日志级别。",
    "is_read": "是否已读。",
    "login_status": "登录结果状态。",
    "login_type": "登录类型。",
    "service_name": "服务名。",
    "label": "展示标签。",
    "score": "置信度或评分。",
    "position": "位置序号，用于标记在正文中的出现顺序。",
    "ref_type": "引用主体类型。",
    "field": "引用字段类型。",
    "owner_id": "被引用主体的业务主键。",
    "reqid": "请求 ID，用于排查问题。",
    "pipeline": "七牛处理队列名。",
    "inputBucket": "七牛输入桶名。",
    "inputKey": "七牛输入对象 key。",
    "desc": "补充描述信息。",
}


SCHEMA_OVERRIDES = {
    ("CreateImageUploadTaskRequest", "hash"): "前端预先计算得到的 qetag/hash，用于上传前预去重。",
    ("CreateImageUploadTaskRequest", "mime_type"): "文件 MIME 类型，例如 image/png。",
    ("CreateImageUploadTaskResponseData", "status"): "图片或任务当前状态，命中预去重时通常可直接使用。",
    ("CompleteImageUploadTaskRequest", "object_key"): "七牛对象 key，应与创建上传任务时返回的 object_key 保持一致。",
    ("UploadTaskStatusResponseData", "status"): "上传任务当前状态，例如 pending、ready、failed。",
    ("SendEmailRequest", "type"): "验证码业务类型，例如注册、找回密码、绑定邮箱或邮箱登录。",
    ("PwdLoginRequest", "username"): "用户名或邮箱，后端会按登录规则识别。",
    ("QQLoginRequest", "code"): "QQ OAuth 授权返回的 code。",
    ("RegisterEmailRequest", "pwd"): "注册时设置的新密码。",
    ("ResetPasswordRequest", "new_password"): "找回密码后要设置的新密码。",
    ("UpdatePasswordRequest", "new_password"): "修改后的新密码。",
    ("UserInfoUpdateRequest", "like_tags"): "用户偏好标签列表，用于个性化展示。",
    ("AdminUserInfoUpdateRequest", "status"): "管理员设置的用户状态值。",
    ("BannerCreateRequest", "href"): "轮播图点击后跳转的链接地址。",
    ("ArticleCreateRequest", "status"): "文章状态，如草稿、待审核或已发布。",
    ("ArticleUpdateRequest", "status"): "更新后的文章状态值。",
    ("ArticleFavoriteRequest", "favor_id"): "目标收藏夹 ID。",
    ("ArticleViewCountRequest", "article_id"): "需要记录浏览的文章 ID。",
    ("FavoriteRequest", "cover"): "收藏夹封面图地址。",
}


REQUEST_EXAMPLES = {
    ("POST", "/api/users/login"): {
        "simple": {
            "summary": "密码登录示例",
            "value": {"username": "testAdmin", "password": "123456"},
        }
    },
    ("POST", "/api/users/email/verify"): {
        "email_login": {
            "summary": "邮箱登录验证码",
            "value": {"type": 4, "email": "user@example.com"},
        },
        "register": {
            "summary": "邮箱注册验证码",
            "value": {"type": 1, "email": "user@example.com"},
        },
    },
    ("POST", "/api/users/email/login"): {
        "simple": {
            "summary": "邮箱验证码登录示例",
            "value": {"email_id": "abc123", "email_code": "123456"},
        }
    },
    ("POST", "/api/images/upload-tasks"): {
        "create_task": {
            "summary": "创建图片上传任务",
            "value": {
                "file_name": "avatar.png",
                "size": 490,
                "mime_type": "image/png",
                "hash": "Fm8TSAox63x45bd-hrHs87ZQPSxx",
            },
        }
    },
    ("POST", "/api/images/upload-tasks/complete"): {
        "manual_complete": {
            "summary": "开发环境手动完成上传",
            "value": {
                "upload_id": 295621418450685952,
                "object_key": "myblogx/images/20260328/Fm8TSAox63x45bd-hrHs87ZQPSxx",
            },
        }
    },
    ("POST", "/api/images/qiniu/callback"): {
        "callback": {
            "summary": "七牛上传完成回调示例",
            "value": {
                "key": "myblogx/images/20260328/Fm8TSAox63x45bd-hrHs87ZQPSxx",
                "hash": "Fm8TSAox63x45bd-hrHs87ZQPSxx",
                "bucket": "myblogx",
                "fsize": 490,
            },
        }
    },
    ("POST", "/api/images/qiniu/audit/callback"): {
        "audit_block": {
            "summary": "七牛审核回调示例",
            "value": {
                "id": "z0.5b8911ea38b9f324a5734c32",
                "pipeline": "0.default",
                "code": 0,
                "desc": "The fop was completed successfully",
                "reqid": "mH0AAOWK5yLQ708V",
                "inputBucket": "myblogx",
                "inputKey": "myblogx/images/20260328/Fm8TSAox63x45bd-hrHs87ZQPSxx",
                "items": [
                    {
                        "cmd": "image-censor/v2/pulp/terror/politician",
                        "code": 0,
                        "desc": "The fop was completed successfully",
                        "result": {
                            "disable": True,
                            "result": {
                                "code": 200,
                                "message": "OK",
                                "suggestion": "block",
                            },
                        },
                    }
                ],
            },
        }
    },
    ("POST", "/api/articles"): {
        "publish_article": {
            "summary": "发布文章示例",
            "value": {
                "title": "测试文章",
                "abstract": "这是一篇用于联调的文章摘要。",
                "content": "# 这是一个标题\\n\\n正文内容。\\n\\n![测试图片](https://REDACTED_CDN_DOMAIN/myblogx/images/20260328/ljZa0YMcc2u6lyAqG-ALnuhewGrY)",
                "category_id": 1,
                "tag_ids": [1, 2],
                "cover": "https://REDACTED_CDN_DOMAIN/myblogx/images/20260328/ljZa0YMcc2u6lyAqG-ALnuhewGrY",
                "comments_toggle": True,
                "status": 2,
            },
        }
    },
}


RESPONSE_EXAMPLES = {
    ("POST", "/api/users/login", "200"): {
        "success": {
            "summary": "登录成功",
            "value": {"code": 0, "msg": "成功", "data": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.xxx.yyy"},
        }
    },
    ("POST", "/api/users/refresh", "200"): {
        "success": {
            "summary": "刷新成功",
            "value": {"code": 0, "msg": "成功", "data": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.new.xxx"},
        }
    },
    ("POST", "/api/users/email/verify", "200"): {
        "success": {
            "summary": "发送成功",
            "value": {"code": 0, "msg": "成功", "data": {"email_id": "abc123", "email": "user@example.com"}},
        }
    },
    ("POST", "/api/users/email/login", "200"): {
        "success": {
            "summary": "登录成功",
            "value": {"code": 0, "msg": "成功", "data": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.xxx.yyy"},
        }
    },
    ("POST", "/api/images/upload-tasks", "200"): {
        "skip_upload": {
            "summary": "命中预去重，直接复用图片",
            "value": {
                "code": 0,
                "msg": "成功",
                "data": {
                    "skip_upload": True,
                    "image_id": 295954303670030336,
                    "status": "ready",
                    "url": "https://REDACTED_CDN_DOMAIN/myblogx/images/20260328/Fm8TSAox63x45bd-hrHs87ZQPSxx",
                    "hash": "Fm8TSAox63x45bd-hrHs87ZQPSxx",
                },
            },
        },
        "need_upload": {
            "summary": "需要继续上传到七牛",
            "value": {
                "code": 0,
                "msg": "成功",
                "data": {
                    "skip_upload": False,
                    "upload_id": 295621418450685952,
                    "provider": "qiniu",
                    "bucket": "myblogx",
                    "object_key": "myblogx/images/20260328/Fm8TSAox63x45bd-hrHs87ZQPSxx",
                    "upload_token": "upload-token",
                    "region": "z2",
                    "expire_at": "2026-03-28T10:00:00+08:00",
                    "max_size": 5242880,
                    "hash": "Fm8TSAox63x45bd-hrHs87ZQPSxx",
                },
            },
        },
    },
    ("GET", "/api/images/upload-tasks/{id}", "200"): {
        "pending": {
            "summary": "等待七牛回调完成",
            "value": {
                "code": 0,
                "msg": "成功",
                "data": {
                    "upload_id": 295621418450685952,
                    "status": "pending",
                    "hash": "Fm8TSAox63x45bd-hrHs87ZQPSxx",
                },
            },
        },
        "ready": {
            "summary": "图片已可用",
            "value": {
                "code": 0,
                "msg": "成功",
                "data": {
                    "upload_id": 295621418450685952,
                    "image_id": 295954303670030336,
                    "status": "ready",
                    "url": "https://REDACTED_CDN_DOMAIN/myblogx/images/20260328/Fm8TSAox63x45bd-hrHs87ZQPSxx",
                    "hash": "Fm8TSAox63x45bd-hrHs87ZQPSxx",
                },
            },
        },
    },
    ("POST", "/api/images/upload-tasks/complete", "200"): {
        "success": {
            "summary": "手动完成上传成功",
            "value": {
                "code": 0,
                "msg": "成功",
                "data": {
                    "upload_id": 295621418450685952,
                    "image_id": 295954303670030336,
                    "status": "ready",
                    "url": "https://REDACTED_CDN_DOMAIN/myblogx/images/20260328/Fm8TSAox63x45bd-hrHs87ZQPSxx",
                },
            },
        }
    },
    ("POST", "/api/images/qiniu/callback", "200"): {
        "success": {
            "summary": "七牛上传回调成功",
            "value": {
                "code": 0,
                "msg": "成功",
                "data": {
                    "upload_id": 295621418450685952,
                    "image_id": 295954303670030336,
                    "status": "ready",
                    "url": "https://REDACTED_CDN_DOMAIN/myblogx/images/20260328/Fm8TSAox63x45bd-hrHs87ZQPSxx",
                },
            },
        }
    },
    ("POST", "/api/images/qiniu/audit/callback", "200"): {
        "success": {
            "summary": "审核结果已接收",
            "value": {"code": 0, "msg": "成功", "data": {}},
        }
    },
    ("POST", "/api/articles", "200"): {
        "success": {
            "summary": "文章发布成功",
            "value": {"code": 0, "msg": "成功", "data": {}},
        }
    },
}


def apply_operation_descriptions(doc: dict) -> None:
    for path, methods in doc["paths"].items():
        for method, op in methods.items():
            if method.startswith("x-"):
                continue
            key = (method.upper(), path)
            if key in OP_OVERRIDES:
                op["description"] = OP_OVERRIDES[key]
                continue
            if str(op.get("description") or "").strip():
                continue
            op["description"] = generic_desc(method.upper(), op.get("summary", path), op)


def apply_parameter_descriptions(doc: dict) -> None:
    for prm in doc.get("components", {}).get("parameters", {}).values():
        if str(prm.get("description") or "").strip():
            continue
        prm["description"] = PARAM_DESC.get(prm.get("name"), f"参数 `{prm.get('name')}` 的作用请结合接口上下文使用。")

    for methods in doc["paths"].values():
        for method, op in methods.items():
            if method.startswith("x-"):
                continue
            for prm in op.get("parameters", []) or []:
                if "$ref" in prm or str(prm.get("description") or "").strip():
                    continue
                name = prm.get("name")
                place = prm.get("in")
                if place == "path":
                    prm["description"] = PARAM_DESC.get(name, f"路径参数 `{name}` 用于定位具体资源。")
                elif place == "query":
                    prm["description"] = PARAM_DESC.get(name, f"查询参数 `{name}` 用于控制筛选、分页或排序。")
                else:
                    prm["description"] = f"参数 `{name}` 的作用请结合接口上下文使用。"


def infer_field_desc(field_name: str, prop: dict) -> str:
    if field_name.endswith("_id"):
        return f"关联的 {field_name[:-3]} ID。"
    if field_name.endswith("_count"):
        return f"{field_name[:-6]} 统计数量。"
    if field_name.endswith("_at"):
        return f"{field_name} 时间。"
    if field_name.endswith("_list"):
        return f"{field_name} 列表数据。"
    if prop.get("type") == "array":
        return f"字段 `{field_name}` 的数组数据。"
    return f"字段 `{field_name}` 的业务含义请结合该接口上下文使用。"


def apply_schema_property_descriptions(doc: dict) -> None:
    for schema_name, schema in doc.get("components", {}).get("schemas", {}).items():
        for field_name, prop in (schema.get("properties") or {}).items():
            if "$ref" in prop or str(prop.get("description") or "").strip():
                continue
            prop["description"] = (
                SCHEMA_OVERRIDES.get((schema_name, field_name))
                or COMMON_FIELD_DESC.get(field_name)
                or infer_field_desc(field_name, prop)
            )


def main() -> None:
    data = json.loads(OPENAPI.read_text(encoding="utf-8-sig"))
    apply_operation_descriptions(data)
    apply_parameter_descriptions(data)
    apply_schema_property_descriptions(data)
    for (method, path), examples in REQUEST_EXAMPLES.items():
        op = data["paths"].get(path, {}).get(method.lower())
        if not op:
            continue
        content = op.get("requestBody", {}).get("content", {}).get("application/json")
        if not content:
            continue
        content.pop("example", None)
        content["examples"] = examples
    for (method, path, status), examples in RESPONSE_EXAMPLES.items():
        op = data["paths"].get(path, {}).get(method.lower())
        if not op:
            continue
        content = op.get("responses", {}).get(status, {}).get("content", {}).get("application/json")
        if not content:
            continue
        content["examples"] = examples
    OPENAPI.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
