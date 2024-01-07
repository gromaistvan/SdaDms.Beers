#pragma warning disable CA1812

using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.IO;
using System.Net.Http;
using System.Runtime.CompilerServices;
using System.Text.Json;
using System.Text.Json.Serialization;
using System.Threading;
using System.Threading.Tasks;
using SixLabors.ImageSharp.Processing;
using Spectre.Console;
using Spectre.Console.Cli;
using Spectre.Console.Rendering;

namespace SdaDms.Beers;

public sealed class DefaultCommand: AsyncCommand<DefaultCommand.Settings>
{
    public sealed class Settings: CommandSettings
    {
        [CommandOption("-s|--service"), DefaultValue("https://api.punkapi.com/v2/beers"), Description("Service URL.")]
        public string ServicePath { get; init; } = null!;

        [CommandOption("-b|--background"), DefaultValue("default"), Description("Background color.")]
        public string BackgroundName { get; init; } = null!;

        public Color Background => Style.TryParse($"default on {BackgroundName}", out Style? style) ? style?.Background ?? Color.Default : Color.Default;

        public Color Foreground => Background.Blend(Color.Yellow, 1f);

        public Uri ServiceUrl => new (ServicePath);
    }

    private sealed record Amount(double Value, string Unit);

    private sealed record Item(string Name, Amount Amount);

    private sealed record Ingredients(IList<Item> Malt);

    private sealed record Beer(int Id)
    {
        public string? Name { get; init; }

        public string? Tagline { get; init; }

        public string? Description { get; init; }

        public double? Ibu { get; init; }

        public Ingredients? Ingredients { get; init; }

        [JsonPropertyName("image_url")]
        public string? Image { get; init; }
    }

    private static readonly JsonSerializerOptions JsonOptions = new () { PropertyNameCaseInsensitive = true };

    private static async Task<List<Beer>> LoadAsync(Settings settings, CancellationToken token = default)
    {
        using HttpClient client = new ();
        Stream response = await client.GetStreamAsync(settings.ServiceUrl, token).ConfigureAwait(false);
        await using ConfiguredAsyncDisposable _ = response.ConfigureAwait(false);
        return await JsonSerializer.DeserializeAsync<List<Beer>>(response, JsonOptions, token).ConfigureAwait(false) ?? [];
    }

    private static async Task<Beer?> ChooseAsync(IList<Beer>? beers, Settings settings, CancellationToken token = default)
    {
        if (beers is null or { Count: 0 }) return null;
        return await
            new SelectionPrompt<Beer>()
            .AddChoices(beers)
            .UseConverter(b => $"{b.Name} ({b.Id})")
            .PageSize(beers.Count)
            .HighlightStyle(new Style(settings.Foreground, settings.Background, Decoration.Bold))
            .ShowAsync(AnsiConsole.Console, token)
            .ConfigureAwait(false);
    }

    private static async Task AlternateScreenAsync(Func<Task> action)
    {
        ControlCode switchScreen = new ("\u001b[?1049h\u001b[H"), returnScreen = new ("\u001b[?1049l");
        AnsiConsole.Write(switchScreen);
        try
        {
            await action.Invoke().ConfigureAwait(false);
        }
        finally
        {
            AnsiConsole.Write(returnScreen);
        }
    }

    private static async Task ShowAsync(Beer beer, Settings settings, CancellationToken token = default)
    {
        Markup nameText = new Markup($"[bold]{beer.Name}[/]").Centered();
        Markup taglineText = new Markup($"{beer.Tagline}").Ellipsis();
        Markup descriptionText = new Markup($"{beer.Description}").Ellipsis();
        BarChart bitternessChart = new BarChart().WithMaxValue(150.0).AddItem(Resources.IBU, beer.Ibu ?? 0.0, settings.Foreground);
        Table maltTable = new Table()
            .AddColumn(Resources.Ingredient)
            .AddColumn(Resources.Amount, c => c.RightAligned())
            .AddColumn(Resources.Unit)
            .Border(TableBorder.Heavy)
            .Expand();
        foreach (Item malt in beer.Ingredients?.Malt ?? [])
        {
            maltTable.AddRow(malt.Name, $"{malt.Amount.Value:0.00}", malt.Amount.Unit);
        }
        using HttpClient client = new ();
        byte[] imageResponse = beer.Image is null ? [] : await client.GetByteArrayAsync(new Uri(beer.Image), token).ConfigureAwait(false);
        IRenderable image = imageResponse.Length == 0 
            ? new Text("")
            : Align.Center(new CanvasImage(imageResponse).Mutate(ctx => ctx.Resize(new ResizeOptions
            {
                Mode = ResizeMode.Max,
                Position = AnchorPositionMode.Center,
                Size = new SixLabors.ImageSharp.Size(AnsiConsole.Profile.Width / 2 - 2, AnsiConsole.Profile.Height - 2)
            })));
        await AlternateScreenAsync(async () =>
        {
            AnsiConsole.Write(
                new Layout().SplitColumns(
                    new Layout().SplitRows(
                        new Layout().SplitRows(
                            CreatePanel(Resources.Name, nameText),
                            CreatePanel(Resources.Tagline, taglineText),
                            CreatePanel(Resources.Description, descriptionText, 3),
                            CreatePanel(Resources.Bitterness, bitternessChart)),
                        CreatePanel(Resources.Malt, maltTable)),
                    CreatePanel(Resources.Image, image)));
            AnsiConsole.Cursor.Hide();
            await AnsiConsole.Console.Input.ReadKeyAsync(true, token).ConfigureAwait(false);
        }).ConfigureAwait(false);
        return;

        Layout CreatePanel(string title, IRenderable content, int ratio = 1)
        {
            var panel = new Panel(content)
            {
                Header = new PanelHeader(title), Padding = new Padding(0),
                Border = BoxBorder.Rounded, BorderStyle = new Style(settings.Foreground, settings.Background, Decoration.Bold),
                Expand = true
            };
            return new Layout(panel) { MinimumSize = 3, Ratio = ratio };
        }
    }

    private static async Task CtrlBreakAsync(Func<CancellationToken, Task> action)
    {
        using CancellationTokenSource tokenSource = new ();
        string title = OperatingSystem.IsWindows() ? Console.Title : "";
        Console.CancelKeyPress += CancelKeyPress;
        try
        {
            if (OperatingSystem.IsWindows())
            {
                Console.Title = Resources.Exit;
            }
            await action.Invoke(tokenSource.Token).ConfigureAwait(false);
        }
        catch (OperationCanceledException)
        { }
        finally
        {
            if (OperatingSystem.IsWindows())
            {
                Console.Title = title;
            }
            Console.CancelKeyPress -= CancelKeyPress;
        }
        return;

        void CancelKeyPress(object? sender, ConsoleCancelEventArgs e)
        {
            #pragma warning disable AccessToDisposedClosure
            tokenSource.Cancel();
            #pragma warning restore AccessToDisposedClosure
            e.Cancel = true;
        }
    }

    private static async Task<int> Main(string[] args) => await new CommandApp<DefaultCommand>().RunAsync(args).ConfigureAwait(false);

    public override async Task<int> ExecuteAsync(CommandContext context, Settings settings)
    {
        ArgumentNullException.ThrowIfNull(context);
        ArgumentNullException.ThrowIfNull(settings);
        await CtrlBreakAsync(async token =>
        {
            IList<Beer> beers = await LoadAsync(settings, token).ConfigureAwait(false);
            for (Beer? beer = await ChooseAsync(beers, settings, token).ConfigureAwait(false);
                 beer is not null;
                 beer = await ChooseAsync(beers, settings, token).ConfigureAwait(false))
            {
                await ShowAsync(beer, settings, token).ConfigureAwait(false);
            }
        }).ConfigureAwait(false);
        return 0;
    }
}
