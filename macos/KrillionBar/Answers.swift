import Foundation

struct Answer: Decodable, Identifiable {
    let answer: String
    let tier: String
    let score: Int
    let quip: String?

    var id: String { answer }
}

struct Round: Identifiable {
    let id: String
    let text: String
    let winners: [Answer]
    let accepted: Int
}

struct DailyResult {
    let date: String
    let dayNumber: Int
    let resetsAt: Date?
    let rounds: [Round]

    var winnerCount: Int { rounds.reduce(0) { $0 + $1.winners.count } }
    var plainText: String {
        rounds.enumerated().map { index, round in
            "\(index + 1). \(round.text)\n" + round.winners.map { "   \($0.answer) (+\($0.score))" }.joined(separator: "\n")
        }.joined(separator: "\n\n")
    }
}

private struct PromptPayload: Decodable {
    let id: String
    let text: String?
    let answers: [Answer]?
}

private struct TodayPayload: Decodable {
    let date: String
    let dayNumber: Int
    let resetsAt: String?
    let prompts: [PromptPayload]

    private enum CodingKeys: String, CodingKey { case date, dayNumber, resetsAt, prompts }

    init(from decoder: Decoder) throws {
        let values = try decoder.container(keyedBy: CodingKeys.self)
        date = try values.decode(String.self, forKey: .date)
        dayNumber = try values.decode(Int.self, forKey: .dayNumber)
        resetsAt = try? values.decode(String.self, forKey: .resetsAt)
        prompts = try values.decode([PromptPayload].self, forKey: .prompts)
    }
}

private struct RevealPayload: Decodable {
    let date: String
    let prompts: [PromptPayload]
}

enum AnswerError: LocalizedError {
    case badResponse(String)
    case changedPuzzle
    case rateLimited
    case denied

    var errorDescription: String? {
        switch self {
        case .badResponse(let detail): "Krillion data changed: \(detail)"
        case .changedPuzzle: "The daily puzzle changed while loading. Try refreshing."
        case .rateLimited: "Krillion is rate limiting requests. Try again later."
        case .denied: "Krillion denied access to today's answers."
        }
    }
}

struct AnswerClient {
    private let baseURL: URL
    private let session: URLSession

    init(baseURL: URL = URL(string: "https://krillion.io")!, session: URLSession? = nil) {
        self.baseURL = baseURL
        if let session {
            self.session = session
        } else {
            let configuration = URLSessionConfiguration.ephemeral
            configuration.httpShouldSetCookies = false
            configuration.httpCookieAcceptPolicy = .never
            self.session = URLSession(configuration: configuration)
        }
    }

    func solve() async throws -> DailyResult {
        do {
            return try await solveOnce()
        } catch AnswerError.changedPuzzle {
            return try await solveOnce()
        }
    }

    private func solveOnce() async throws -> DailyResult {
        let today: TodayPayload = try await get("/api/today")
        guard today.date.range(of: #"^\d{4}-\d{2}-\d{2}$"#, options: .regularExpression) != nil,
              Self.dateFormatter.date(from: today.date) != nil else {
            throw AnswerError.badResponse("invalid puzzle date")
        }
        guard today.prompts.count == 7 else {
            throw AnswerError.badResponse("expected seven prompts")
        }
        let reveal: RevealPayload = try await get("/api/reveal?date=\(today.date)")
        guard reveal.date == today.date else { throw AnswerError.changedPuzzle }

        var byID: [String: PromptPayload] = [:]
        for prompt in reveal.prompts {
            guard !prompt.id.isEmpty, byID.updateValue(prompt, forKey: prompt.id) == nil else {
                throw AnswerError.badResponse("duplicate or empty reveal ID")
            }
        }

        var seen = Set<String>()
        let rounds = try today.prompts.map { prompt -> Round in
            guard !prompt.id.isEmpty,
                  seen.insert(prompt.id).inserted,
                  let text = prompt.text, !text.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty,
                  let revealed = byID[prompt.id],
                  let answers = revealed.answers else {
                throw AnswerError.badResponse("incomplete prompt data")
            }
            let winners = answers.filter { $0.tier == "krillion" }
            guard !winners.isEmpty,
                  winners.allSatisfy({ !$0.answer.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty && $0.score == 100 }) else {
                throw AnswerError.badResponse("missing or invalid top answer")
            }
            return Round(id: prompt.id, text: text, winners: winners, accepted: answers.count)
        }

        return DailyResult(
            date: today.date,
            dayNumber: today.dayNumber,
            resetsAt: today.resetsAt.flatMap(Self.parseReset),
            rounds: rounds
        )
    }

    private func get<T: Decodable>(_ path: String) async throws -> T {
        var request = URLRequest(url: URL(string: path, relativeTo: baseURL)!)
        request.timeoutInterval = 12
        request.setValue("application/json", forHTTPHeaderField: "Accept")
        request.setValue("krillion-solver-macos/0.1", forHTTPHeaderField: "User-Agent")
        let (data, response) = try await session.data(for: request)
        guard let response = response as? HTTPURLResponse else {
            throw AnswerError.badResponse("missing HTTP response")
        }
        switch response.statusCode {
        case 200: break
        case 429: throw AnswerError.rateLimited
        case 401, 403: throw AnswerError.denied
        default: throw AnswerError.badResponse("HTTP \(response.statusCode)")
        }
        guard data.count <= 8 * 1024 * 1024 else {
            throw AnswerError.badResponse("response exceeded 8 MiB")
        }
        do {
            return try JSONDecoder().decode(T.self, from: data)
        } catch {
            throw AnswerError.badResponse("unexpected JSON")
        }
    }

    private static let dateFormatter: DateFormatter = {
        let formatter = DateFormatter()
        formatter.locale = Locale(identifier: "en_US_POSIX")
        formatter.timeZone = TimeZone(secondsFromGMT: 0)
        formatter.dateFormat = "yyyy-MM-dd"
        formatter.isLenient = false
        return formatter
    }()

    private static func parseReset(_ value: String) -> Date? {
        let formatter = ISO8601DateFormatter()
        return formatter.date(from: value)
    }
}
