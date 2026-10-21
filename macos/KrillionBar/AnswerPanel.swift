import SwiftUI

private enum Palette {
    static let ocean = Color(red: 6/255, green: 22/255, blue: 35/255)
    static let surface = Color(red: 12/255, green: 36/255, blue: 51/255)
    static let pearl = Color(red: 228/255, green: 237/255, blue: 241/255)
    static let muted = Color(red: 120/255, green: 149/255, blue: 168/255)
    static let cyan = Color(red: 86/255, green: 216/255, blue: 237/255)
    static let gold = Color(red: 244/255, green: 201/255, blue: 93/255)
    static let coral = Color(red: 240/255, green: 100/255, blue: 121/255)
    static let line = Color(red: 36/255, green: 67/255, blue: 84/255)
}

struct AnswerPanel: View {
    @ObservedObject var store: AnswerStore

    var body: some View {
        VStack(spacing: 0) {
            header
            Rectangle().fill(Palette.line).frame(height: 1)

            Group {
                if store.loading {
                    stateView(symbol: "water.waves", title: "Descending into today's dive…", detail: "Finding the best answers.")
                } else if let error = store.error {
                    VStack(spacing: 12) {
                        stateView(symbol: "exclamationmark.circle", title: "The current got away.", detail: error)
                        Button("Try again") { store.refresh() }
                            .buttonStyle(.borderedProminent)
                    }
                } else if let result = store.result {
                    results(result)
                } else {
                    stateView(symbol: "water.waves", title: "Ready to dive", detail: "Click Refresh to load today's answers.")
                }
            }
            .frame(maxWidth: .infinity)

            Rectangle().fill(Palette.line).frame(height: 1)
            footer
        }
        .background(Palette.ocean)
        .foregroundStyle(Palette.pearl)
        .preferredColorScheme(.dark)
    }

    private var header: some View {
        HStack(alignment: .center, spacing: 12) {
            Image("Shrimp")
                .resizable()
                .renderingMode(.template)
                .scaledToFit()
                .frame(width: 28, height: 28)
                .foregroundStyle(Palette.coral)
                .accessibilityHidden(true)
            VStack(alignment: .leading, spacing: 3) {
                Text("KRILLION SOLVER")
                    .font(.system(size: 13, weight: .bold, design: .monospaced))
                    .tracking(1.5)
                    .foregroundStyle(Palette.coral)
                Text("THE DAILY CATCH")
                    .font(.system(size: 10, weight: .medium, design: .monospaced))
                    .tracking(1.2)
                    .foregroundStyle(Palette.muted)
            }
            Spacer()
            Button { store.refresh() } label: {
                Image(systemName: "arrow.clockwise")
                    .font(.system(size: 13, weight: .semibold))
            }
            .buttonStyle(.plain)
            .foregroundStyle(Palette.cyan)
            .disabled(store.loading)
            .help("Refresh answers")
            .accessibilityLabel("Refresh answers")
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 13)
    }

    private func results(_ result: DailyResult) -> some View {
        VStack(spacing: 0) {
            HStack {
                Text(String(format: "DIVE %03d", result.dayNumber))
                Spacer()
                Text(result.date)
                Text("·")
                Text("\(result.winnerCount) winners")
            }
            .font(.system(size: 10, weight: .medium, design: .monospaced))
            .foregroundStyle(Palette.muted)
            .padding(.horizontal, 16)
            .padding(.vertical, 11)

            ScrollView {
                VStack(spacing: 7) {
                    ForEach(Array(result.rounds.enumerated()), id: \.element.id) { index, round in
                        RoundCard(number: index + 1, round: round, copy: store.copy)
                    }
                }
                .padding(.horizontal, 12)
                .padding(.bottom, 12)
            }
            .frame(maxHeight: 468)

            HStack {
                Text("700 POSSIBLE POINTS")
                    .font(.system(size: 10, weight: .semibold, design: .monospaced))
                    .foregroundStyle(Palette.cyan)
                Spacer()
                Button("COPY ALL") { store.copy(result.plainText) }
                    .font(.system(size: 10, weight: .bold, design: .monospaced))
                    .buttonStyle(.plain)
                    .foregroundStyle(Palette.gold)
                    .accessibilityLabel("Copy all answers")
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 11)
        }
    }

    private func stateView(symbol: String, title: String, detail: String) -> some View {
        VStack(spacing: 10) {
            Image(systemName: symbol)
                .font(.system(size: 28, weight: .light))
                .foregroundStyle(Palette.cyan)
            Text(title)
                .font(.system(size: 13, weight: .semibold))
                .foregroundStyle(Palette.pearl)
            Text(detail)
                .font(.system(size: 11))
                .foregroundStyle(Palette.muted)
                .multilineTextAlignment(.center)
        }
        .padding(24)
        .frame(height: 220)
    }

    private var footer: some View {
        HStack {
            Link("OPEN KRILLION", destination: URL(string: "https://krillion.io")!)
                .foregroundStyle(Palette.cyan)
            Spacer()
            Button("QUIT") { NSApplication.shared.terminate(nil) }
                .foregroundStyle(Palette.muted)
        }
        .font(.system(size: 10, weight: .semibold, design: .monospaced))
        .buttonStyle(.plain)
        .padding(.horizontal, 16)
        .padding(.vertical, 11)
    }
}

private struct RoundCard: View {
    let number: Int
    let round: Round
    let copy: (String) -> Void
    @State private var expanded = false

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            HStack(spacing: 10) {
                Text(String(format: "%02d", number))
                    .font(.system(size: 11, weight: .medium, design: .monospaced))
                    .foregroundStyle(Palette.cyan)
                Button { expanded.toggle() } label: {
                    HStack(spacing: 6) {
                        Text(round.winners[0].answer)
                            .font(.system(size: 13, weight: .semibold))
                            .lineLimit(1)
                            .foregroundStyle(Palette.gold)
                        if round.winners.count > 1 {
                            Text("+\(round.winners.count - 1)")
                                .font(.system(size: 10, weight: .medium, design: .monospaced))
                                .foregroundStyle(Palette.cyan)
                        }
                        Spacer(minLength: 0)
                        Image(systemName: expanded ? "chevron.up" : "chevron.down")
                            .font(.system(size: 9, weight: .bold))
                            .foregroundStyle(Palette.muted)
                    }
                    .contentShape(Rectangle())
                }
                .buttonStyle(.plain)
                .accessibilityLabel("Round \(number), \(round.winners[0].answer), details")

                Button { copy(round.winners[0].answer) } label: {
                    Image(systemName: "doc.on.doc")
                        .font(.system(size: 11))
                }
                .buttonStyle(.plain)
                .foregroundStyle(Palette.muted)
                .help("Copy \(round.winners[0].answer)")
                .accessibilityLabel("Copy \(round.winners[0].answer)")
            }
            if expanded {
                VStack(alignment: .leading, spacing: 8) {
                    Text(round.text)
                        .foregroundStyle(Palette.pearl)
                    if let quip = round.winners[0].quip, !quip.isEmpty {
                        Text(quip)
                            .foregroundStyle(Palette.muted)
                    }
                    ForEach(Array(round.winners.dropFirst().enumerated()), id: \.offset) { _, alternative in
                        VStack(alignment: .leading, spacing: 4) {
                            HStack {
                                Text("ALSO  \(alternative.answer)")
                                    .foregroundStyle(Palette.cyan)
                                Spacer()
                                Button { copy(alternative.answer) } label: {
                                    Image(systemName: "doc.on.doc")
                                }
                                .buttonStyle(.plain)
                                .foregroundStyle(Palette.muted)
                                .help("Copy \(alternative.answer)")
                            }
                            if let quip = alternative.quip, !quip.isEmpty {
                                Text(quip).foregroundStyle(Palette.muted)
                            }
                        }
                    }
                    Text("\(round.accepted.formatted()) accepted answers")
                        .foregroundStyle(Palette.muted)
                }
                .font(.system(size: 11))
                .padding(.leading, 31)
                .padding(.top, 10)
            }
        }
        .padding(.horizontal, 11)
        .padding(.vertical, 11)
        .background(Palette.surface, in: RoundedRectangle(cornerRadius: 8))
        .overlay(RoundedRectangle(cornerRadius: 8).stroke(Palette.line, lineWidth: 1))
    }
}
