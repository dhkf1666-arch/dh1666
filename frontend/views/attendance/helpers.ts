export function getCurrentYearMonth() {
  const now = new Date();
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}`;
}

export function getToday() {
  return new Date().toISOString().split("T")[0];
}

export function formatDate(date: string | Date | null | undefined) {
  if (!date) return "-";
  const str = String(date);
  if (str.includes("-")) return str.substring(0, 10);
  const d = new Date(date);
  if (isNaN(d.getTime())) return "-";
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(
    d.getDate(),
  ).padStart(2, "0")}`;
}

export function formatDateShort(dateStr: string | Date | null | undefined) {
  if (!dateStr) return "";
  if (typeof dateStr === "string" && dateStr.includes("-")) {
    return dateStr.slice(5);
  }
  if (dateStr instanceof Date) {
    const month = String(dateStr.getMonth() + 1).padStart(2, "0");
    const day = String(dateStr.getDate()).padStart(2, "0");
    return `${month}-${day}`;
  }
  return String(dateStr);
}

export function getAttendanceRateType(rate: number) {
  if (rate >= 90) return "success";
  if (rate >= 70) return "warning";
  return "danger";
}

export function getStatusClass(status: string) {
  const map: Record<string, string> = {
    work: "status-work",
    rest_half: "status-rest-half",
    rest_full: "status-rest-full",
    leave: "status-leave",
    absent: "status-absent",
    off_post: "status-off-post",
    resigned: "status-resigned",
  };
  return map[status] || "";
}

export function getStatusText(status: string) {
  const map: Record<string, string> = {
    work: "出勤",
    rest_half: "半休",
    rest_full: "全休",
    leave: "半假",
    absent: "旷工",
    off_post: "全假",
    resigned: "离职",
    "": "-",
  };
  return map[status] || "-";
}

export function getPenaltyCategoryType(category: string) {
  const map: Record<string, string> = {
    迟到: "warning",
    早退: "warning",
    小厕超时: "danger",
    大厕超时: "danger",
    吃饭超时: "danger",
    抽烟或休息超时: "danger",
    旷工: "danger",
    其他: "info",
  };
  return map[category] || "info";
}

export function getScoreType(score: number) {
  if (score >= 15) return "success"; // 15分以上 - 卓越（绿色）
  if (score >= 12) return "primary"; // 12-14.9分 - 优秀（蓝色）
  if (score === 10) return "success"; // 10分 - 良好（绿色）
  if (score >= 7) return "warning"; // 7-9.9分 - 合格（橙色）
  if (score >= 6) return "danger"; // 6-6.9分 - 待提升（红色）
  return "danger"; // 6分以下 - 不合格（深红色）
}

export function getGradeType(grade: string) {
  const map: Record<string, string> = {
    优秀: "success", // 绿色
    满分: "success", // 绿色（也可以改为 primary 蓝色）
    良好: "primary", // 蓝色
    合格: "warning", // 橙色
    待提升: "danger", // 红色
    不合格: "danger", // 红色
  };
  return map[grade] || "info";
}

export function getScoreRecordOperator(record: any): string {
  const name = record?.operator || record?.operatorName;
  return name && String(name).trim() ? String(name).trim() : "-";
}

export function getPenaltyCreatorName(record: any): string {
  const name =
    record?.createdByName ||
    record?.created_by_name ||
    record?.created_by ||
    record?.createdBy;
  return name && String(name).trim() ? String(name).trim() : "-";
}
