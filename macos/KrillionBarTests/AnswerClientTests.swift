import Foundation
import XCTest
@testable import KrillionBar

private final class StubProtocol: URLProtocol {
    static var handler: ((URLRequest) throws -> (Int, Data))?

    override class func canInit(with request: URLRequest) -> Bool { true }
    override class func canonicalRequest(for request: URLRequest) -> URLRequest { request }

    override func startLoading() {
        do {
            let (status, data) = try Self.handler!(request)
            let response = HTTPURLResponse(url: request.url!, statusCode: status, httpVersion: "HTTP/1.1", headerFields: ["Content-Type": "application/json"])!
            client?.urlProtocol(self, didReceive: response, cacheStoragePolicy: .notAllowed)
            client?.urlProtocol(self, didLoad: data)
            client?.urlProtocolDidFinishLoading(self)
        } catch {
            client?.urlProtocol(self, didFailWithError: error)
        }
    }

    override func stopLoading() {}
}

final class AnswerClientTests: XCTestCase {
    private func client() -> AnswerClient {
        let configuration = URLSessionConfiguration.ephemeral
        configuration.protocolClasses = [StubProtocol.self]
        return AnswerClient(baseURL: URL(string: "https://krillion.test")!, session: URLSession(configuration: configuration))
    }

    private func fixtures() throws -> (Data, Data) {
        let todayPrompts = (1...7).map { ["id": "p\($0)", "text": "Prompt \($0)"] }
        let revealPrompts: [[String: Any]] = (1...7).reversed().map { number in
            let winners: [[String: Any]] = (number == 4 ? ["First", "Second"] : ["Winner \(number)"]).map {
                ["answer": $0, "tier": "krillion", "score": 100, "quip": "Detail"]
            }
            return ["id": "p\(number)", "answers": [["answer": "Common", "tier": "schooler", "score": 30]] + winners]
        }
        return (
            try JSONSerialization.data(withJSONObject: ["date": "2026-10-21", "dayNumber": 98, "resetsAt": "2026-10-22T04:00:00Z", "prompts": todayPrompts]),
            try JSONSerialization.data(withJSONObject: ["date": "2026-10-21", "prompts": revealPrompts])
        )
    }

    func testMatchesRoundsByIDAndKeepsEveryTopAnswer() async throws {
        let (today, reveal) = try fixtures()
        StubProtocol.handler = { request in
            XCTAssertEqual(request.httpMethod, "GET")
            XCTAssertNil(request.value(forHTTPHeaderField: "Cookie"))
            if request.url?.path == "/api/today" { return (200, today) }
            XCTAssertEqual(request.url?.query, "date=2026-10-21")
            return (200, reveal)
        }
        let result = try await client().solve()
        XCTAssertEqual(result.rounds.count, 7)
        XCTAssertEqual(result.rounds[0].winners[0].answer, "Winner 1")
        XCTAssertEqual(result.rounds[3].winners.map(\.answer), ["First", "Second"])
        XCTAssertEqual(result.winnerCount, 8)
        XCTAssertTrue(result.plainText.contains("Second (+100)"))
        XCTAssertNotNil(result.resetsAt)
    }

    func testChangedPuzzleRetriesOnlyOnce() async throws {
        let (today, _) = try fixtures()
        var calls = 0
        let wrong = try JSONSerialization.data(withJSONObject: ["date": "2026-10-20", "prompts": []])
        StubProtocol.handler = { request in
            calls += 1
            return request.url?.path == "/api/today" ? (200, today) : (200, wrong)
        }
        do {
            _ = try await client().solve()
            XCTFail("Expected a date mismatch")
        } catch AnswerError.changedPuzzle {
            XCTAssertEqual(calls, 4)
        }
    }

    func testRejectsInvalidWinner() async throws {
        let (today, _) = try fixtures()
        let prompts: [[String: Any]] = (1...7).map { number in
            ["id": "p\(number)", "answers": [["answer": "Wrong", "tier": "krillion", "score": 90]]]
        }
        let reveal = try JSONSerialization.data(withJSONObject: ["date": "2026-10-21", "prompts": prompts])
        StubProtocol.handler = { request in
            request.url?.path == "/api/today" ? (200, today) : (200, reveal)
        }
        do {
            _ = try await client().solve()
            XCTFail("Expected validation failure")
        } catch AnswerError.badResponse { }
    }
}
