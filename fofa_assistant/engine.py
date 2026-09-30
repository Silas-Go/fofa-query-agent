"""Agent pipeline: natural language -> condition JSON -> validation -> query."""

import logging
import re

from .generator import generate_query
from .parser import parse_conditions
from .validator import result_status, validate_conditions

STATUSES = ("成功", "输入非法", "条件矛盾", "需要澄清", "能力不支持", "依赖失败", "暂未支持")
SOURCE = {"来源": "https://fofa.info/", "依据": "FOFA 官方搜索语法表；结构化条件通过本地校验后生成，未执行资产查询。", "核对日期": "2026-09-30"}


def validate_batch(data):
    if not isinstance(data, list):
        raise ValueError("输入必须是 JSON 数组。")
    if not 1 <= len(data) <= 1000:
        raise ValueError("每批需要 1–1000 道题。")
    seen = set()
    for i, item in enumerate(data, 1):
        if not isinstance(item, dict):
            raise ValueError(f"第 {i} 条必须是对象。")
        for key in ("题号", "自然语言输入"):
            if not isinstance(item.get(key), str) or not item[key].strip():
                raise ValueError(f"第 {i} 条的“{key}”必须是非空字符串。")
        if len(item["自然语言输入"]) > 20000:
            raise ValueError(f"第 {i} 条题目超过 20000 字符。")
        if item["题号"] in seen:
            raise ValueError(f"题号重复：{item['题号']}")
        seen.add(item["题号"])


def answer(item):
    result = {"题号": item["题号"], "自然语言输入": item["自然语言输入"],
              "状态": "依赖失败", "查询语句": None, "说明": "", "依据": [],
              "结构化条件": None, "校验结果": {"通过": False, "错误": []}}
    try:
        document = parse_conditions(item["自然语言输入"])
        result["结构化条件"] = document
        validation = validate_conditions(document)
        result["状态"] = result_status(validation)
        result["校验结果"] = {"通过": validation["valid"], "错误": validation["errors"]}
        if not validation["valid"]:
            result["说明"] = "；".join(e["message"] for e in validation["errors"])
            return result
        result["结构化条件"] = validation["normalized"]
        result["查询语句"] = generate_query(result["结构化条件"])
        result["说明"] = "结构化条件已通过本地校验；查询由条件树生成，未执行 FOFA 搜索。"
        if re.search(r"B\s*段", item["自然语言输入"], re.I):
            result["说明"] += "“B 段”按 /16 网段解释。"
        result["依据"] = [dict(SOURCE)]
        return result
    except Exception:
        logging.getLogger(__name__).exception("Question processing failed: %s", item["题号"])
        result.update({"状态": "依赖失败", "查询语句": None, "说明": "本题处理失败；其他题目可继续。",
                       "校验结果": {"通过": False, "错误": [{"code": "INTERNAL_ERROR", "path": "$", "message": "处理过程发生内部错误。", "status": "依赖失败"}]}})
        return result


def answer_batch(data):
    validate_batch(data)
    return [answer(item) for item in data]
