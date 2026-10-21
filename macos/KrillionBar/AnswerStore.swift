import AppKit
import SwiftUI

@MainActor
final class AnswerStore: ObservableObject {
    @Published private(set) var result: DailyResult?
    @Published private(set) var loading = false
    @Published private(set) var error: String?

    private let client = AnswerClient()
    private var fetchedAt: Date?

    func loadIfNeeded() {
        guard !loading else { return }
        if let result, let fetchedAt,
           Date().timeIntervalSince(fetchedAt) < 15 * 60,
           result.resetsAt.map({ Date() < $0 }) ?? true {
            return
        }
        refresh()
    }

    func refresh() {
        guard !loading else { return }
        result = nil
        error = nil
        loading = true
        Task {
            do {
                result = try await client.solve()
                fetchedAt = Date()
            } catch {
                self.error = error.localizedDescription
            }
            loading = false
        }
    }

    func copy(_ text: String) {
        NSPasteboard.general.clearContents()
        NSPasteboard.general.setString(text, forType: .string)
    }
}
