import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";

interface AccessReview {
  id: string;
  org_id: string;
  title: string;
  status: "PENDING" | "IN_PROGRESS" | "COMPLETED";
  due_date: string;
  total_users: number;
  reviewed_users: number;
  created_by: string;
  created_at: string;
  review_items?: ReviewItem[];
}

interface ReviewItem {
  id: string;
  review_id: string;
  user_id: string;
  role_id: string;
  current_status: "PENDING" | "APPROVED" | "REVOKED";
  reviewer_id?: string;
  reviewed_at?: string;
  notes?: string;
  user?: { id: string; email: string; name: string; status: string };
  role?: { id: string; name: string };
}

const API_BASE = import.meta.env.VITE_API_URL || "/api/v1";

export default function AccessReview() {
  const [reviews, setReviews] = useState<AccessReview[]>([]);
  const [selectedReview, setSelectedReview] = useState<AccessReview | null>(
    null,
  );
  const [loading, setLoading] = useState(true);
  const [uiStrings, setUiStrings] = useState<any>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newReview, setNewReview] = useState({ title: "", due_date: "" });

  useEffect(() => {
    fetch("/api/v1/ui/global_config")
      .then((res) => res.json())
      .then((data) =>
        setUiStrings(
          data.ui_strings?.security_v2 ||
            data.system_metadata?.ui_strings?.security_v2,
        ),
      )
      .catch((err) => console.error("Failed to fetch UI strings", err));
  }, []);
  const [searchParams] = useSearchParams();
  const orgId = searchParams.get("org_id") || "";

  useEffect(() => {
    fetchReviews();
  }, [orgId]);

  const fetchReviews = async () => {
    setLoading(true);
    try {
      const res = await fetch(
        `${API_BASE}/security/access-reviews?org_id=${orgId}`,
      );
      const data = await res.json();
      setReviews(data.reviews || []);
    } catch (err) {
      console.error("Failed to fetch access reviews:", err);
    }
    setLoading(false);
  };

  const fetchReviewDetails = async (id: string) => {
    try {
      const res = await fetch(`${API_BASE}/security/access-reviews/${id}`);
      const data = await res.json();
      setSelectedReview(data);
    } catch (err) {
      console.error("Failed to fetch review details:", err);
    }
  };

  const handleCreateReview = async () => {
    try {
      await fetch(`${API_BASE}/security/access-reviews?org_id=${orgId}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(newReview),
      });
      setShowCreateModal(false);
      setNewReview({ title: "", due_date: "" });
      fetchReviews();
    } catch (err) {
      console.error("Failed to create access review:", err);
    }
  };

  const handleReviewAction = async (
    itemId: string,
    action: "APPROVED" | "REVOKED",
    notes: string = "",
  ) => {
    try {
      await fetch(`${API_BASE}/security/review-items/update`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ item_id: itemId, status: action, notes }),
      });
      if (selectedReview) {
        fetchReviewDetails(selectedReview.id);
      }
    } catch (err) {
      console.error("Failed to update review item:", err);
    }
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString();
  };

  const getStatusBadge = (status: string) => {
    const styles: Record<string, string> = {
      PENDING: "bg-yellow-600/20 text-yellow-400",
      IN_PROGRESS: "bg-blue-600/20 text-blue-400",
      COMPLETED: "bg-green-600/20 text-green-400",
      APPROVED: "bg-green-600/20 text-green-400",
      REVOKED: "bg-red-600/20 text-red-400",
    };
    return (
      <span
        className={`px-2 py-1 rounded text-sm ${styles[status] || "bg-gray-600/20 text-gray-400"}`}
      >
        {uiStrings?.[status.toLowerCase() + "_label"] || status}
      </span>
    );
  };

  const getProgress = (review: AccessReview) => {
    if (review.total_users === 0) return 0;
    return Math.round((review.reviewed_users / review.total_users) * 100);
  };

  return (
    <div className="p-6 bg-gray-900 min-h-screen text-white">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold">
            {uiStrings?.access_reviews_title || "Access Reviews"}
          </h1>
          <p className="text-gray-400">
            {uiStrings?.access_reviews_subtitle ||
              "Periodic review of user access and role assignments"}
          </p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg transition-colors flex items-center gap-2"
        >
          <span>+</span> {uiStrings?.start_new_review_btn || "Start New Review"}
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-gray-800/50 backdrop-blur-lg rounded-xl p-4 border border-gray-700">
          <div className="text-3xl font-bold text-blue-400">
            {reviews.length}
          </div>
          <div className="text-gray-400 text-sm">
            {uiStrings?.stat_total_reviews || "Total Reviews"}
          </div>
        </div>
        <div className="bg-gray-800/50 backdrop-blur-lg rounded-xl p-4 border border-gray-700">
          <div className="text-3xl font-bold text-yellow-400">
            {reviews.filter((r) => r.status === "IN_PROGRESS").length}
          </div>
          <div className="text-gray-400 text-sm">
            {uiStrings?.stat_in_progress || "In Progress"}
          </div>
        </div>
        <div className="bg-gray-800/50 backdrop-blur-lg rounded-xl p-4 border border-gray-700">
          <div className="text-3xl font-bold text-green-400">
            {reviews.filter((r) => r.status === "COMPLETED").length}
          </div>
          <div className="text-gray-400 text-sm">
            {uiStrings?.stat_completed || "Completed"}
          </div>
        </div>
        <div className="bg-gray-800/50 backdrop-blur-lg rounded-xl p-4 border border-gray-700">
          <div className="text-3xl font-bold text-red-400">
            {
              reviews.filter((r) => {
                const due = new Date(r.due_date);
                return due < new Date() && r.status !== "COMPLETED";
              }).length
            }
          </div>
          <div className="text-gray-400 text-sm">
            {uiStrings?.stat_overdue || "Overdue"}
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Reviews List */}
        <div className="bg-gray-800/50 backdrop-blur-lg rounded-xl border border-gray-700 overflow-hidden">
          <div className="px-4 py-3 bg-gray-700/50 border-b border-gray-700">
            <h2 className="font-semibold">
              {uiStrings?.all_reviews_title || "All Reviews"}
            </h2>
          </div>
          {loading ? (
            <div className="flex justify-center items-center h-48">
              <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-blue-500"></div>
            </div>
          ) : (
            <div className="divide-y divide-gray-700">
              {reviews.map((review) => (
                <div
                  key={review.id}
                  onClick={() => fetchReviewDetails(review.id)}
                  className={`p-4 cursor-pointer hover:bg-gray-700/30 transition-colors ${
                    selectedReview?.id === review.id ? "bg-gray-700/50" : ""
                  }`}
                >
                  <div className="flex justify-between items-start mb-2">
                    <h3 className="font-medium">{review.title}</h3>
                    {getStatusBadge(review.status)}
                  </div>
                  <div className="text-sm text-gray-400 mb-2">
                    Due: {formatDate(review.due_date)}
                  </div>
                  <div className="w-full bg-gray-700 rounded-full h-2">
                    <div
                      className="bg-blue-500 h-2 rounded-full transition-all"
                      style={{ width: `${getProgress(review)}%` }}
                    ></div>
                  </div>
                  <div className="text-xs text-gray-400 mt-1">
                    {review.reviewed_users} / {review.total_users} reviewed
                  </div>
                </div>
              ))}
              {reviews.length === 0 && (
                <div className="p-8 text-center text-gray-400">
                  {uiStrings?.no_reviews_found ||
                    "No access reviews found. Start a new review to begin."}
                </div>
              )}
            </div>
          )}
        </div>

        {/* Review Details */}
        <div className="bg-gray-800/50 backdrop-blur-lg rounded-xl border border-gray-700 overflow-hidden">
          <div className="px-4 py-3 bg-gray-700/50 border-b border-gray-700">
            <h2 className="font-semibold">
              {selectedReview
                ? selectedReview.title
                : uiStrings?.select_review_placeholder || "Select a Review"}
            </h2>
          </div>
          {selectedReview ? (
            <div className="p-4 max-h-[500px] overflow-y-auto">
              {selectedReview.review_items?.map((item) => (
                <div
                  key={item.id}
                  className="p-3 mb-2 bg-gray-700/30 rounded-lg border border-gray-600"
                >
                  <div className="flex justify-between items-start mb-2">
                    <div>
                      <div className="font-medium">
                        {item.user?.name ||
                          uiStrings?.unknown_user ||
                          "Unknown User"}
                      </div>
                      <div className="text-sm text-gray-400">
                        {item.user?.email}
                      </div>
                    </div>
                    {getStatusBadge(item.current_status)}
                  </div>
                  <div className="text-sm text-gray-400 mb-3">
                    {uiStrings?.role_label || "Role:"}{" "}
                    <span className="text-purple-400">
                      {item.role?.name || uiStrings?.unknown_label || "Unknown"}
                    </span>
                  </div>
                  {item.current_status === "PENDING" && (
                    <div className="flex gap-2">
                      <button
                        onClick={() => handleReviewAction(item.id, "APPROVED")}
                        className="flex-1 px-3 py-1 bg-green-600 hover:bg-green-700 rounded text-sm"
                      >
                        {uiStrings?.approve_btn || "Approve"}
                      </button>
                      <button
                        onClick={() => handleReviewAction(item.id, "REVOKED")}
                        className="flex-1 px-3 py-1 bg-red-600 hover:bg-red-700 rounded text-sm"
                      >
                        {uiStrings?.revoke_btn || "Revoke"}
                      </button>
                    </div>
                  )}
                </div>
              ))}
              {(!selectedReview.review_items ||
                selectedReview.review_items.length === 0) && (
                <div className="text-center text-gray-400 py-8">
                  {uiStrings?.no_users_to_review ||
                    "No users to review in this cycle"}
                </div>
              )}
            </div>
          ) : (
            <div className="p-8 text-center text-gray-400">
              {uiStrings?.click_review_hint ||
                "Click on a review to see details"}
            </div>
          )}
        </div>
      </div>

      {/* Create Review Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-xl p-6 w-full max-w-md border border-gray-700">
            <h2 className="text-xl font-bold mb-4">
              {uiStrings?.modal_title_new_review || "Start New Access Review"}
            </h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm text-gray-400 mb-1">
                  Review Title
                </label>
                <input
                  type="text"
                  value={newReview.title}
                  onChange={(e) =>
                    setNewReview({ ...newReview, title: e.target.value })
                  }
                  placeholder={
                    uiStrings?.placeholder_review_title ||
                    "e.g., Q1 2025 Access Review"
                  }
                  className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg focus:outline-none focus:border-blue-500"
                />
              </div>
              <div>
                <label className="block text-sm text-gray-400 mb-1">
                  Due Date
                </label>
                <input
                  type="date"
                  value={newReview.due_date}
                  onChange={(e) =>
                    setNewReview({ ...newReview, due_date: e.target.value })
                  }
                  className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg focus:outline-none focus:border-blue-500"
                />
              </div>
            </div>
            <div className="flex justify-end gap-3 mt-6">
              <button
                onClick={() => setShowCreateModal(false)}
                className="px-4 py-2 text-gray-400 hover:text-white"
              >
                {uiStrings?.cancel_btn || "Cancel"}
              </button>
              <button
                onClick={handleCreateReview}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg"
              >
                {uiStrings?.create_review_btn || "Create Review"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
