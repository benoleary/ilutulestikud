import { Component } from '@angular/core';
import { OnInit, OnDestroy } from '@angular/core/src/metadata/lifecycle_hooks';
import { MatDialog } from '@angular/material';
import { MatTabChangeEvent } from '@angular/material';
import { Observable } from 'rxjs/Rx';
import { Subscription } from 'rxjs/Subscription';
import { IlutulestikudService } from './ilutulestikud.service';
import { BackendIdentification } from './models/backendidentification.model';
import { Player } from './models/player.model'
import { Ruleset } from './models/ruleset.model'
import { TurnSummary } from './models/turnsummary.model'
import { AddPlayerDialogComponent } from './components/addplayerdialog.component'
import { CreateGameDialogComponent } from './components/creategamedialog.component'

@Component({
  selector: 'app-ilutulestikud',
  templateUrl: './ilutulestikud.component.html',
  styleUrls: ['ilutulestikud.component.css']
})
export class IlutulestikudComponent implements OnInit, OnDestroy
{
  informationText: string;
  selectedPlayer: Player;
  registeredPlayers: Player[];
  availableColors: string[];
  availableRulesets: Ruleset[];
  turnSummariesOfGamesWithPlayer: TurnSummary[];
  gameTurnSummariesSubscription: Subscription;
  isAwaitingGameTurnSummaries: boolean;
  selectedGameTabIndex: number;
  selectedGameIdentification: BackendIdentification;
  

  constructor(public ilutulestikudService: IlutulestikudService, public materialDialog: MatDialog)
  {
    this.selectedPlayer = null;
    this.registeredPlayers = [];
    this.availableColors = [];
    this.availableRulesets = [];
    this.informationText = null;
    this.turnSummariesOfGamesWithPlayer = [];
    this.gameTurnSummariesSubscription = null;
    this.isAwaitingGameTurnSummaries = false;
    this.selectedGameTabIndex = 0;
    this.selectedGameIdentification = null;
  }

  ngOnInit(): void
  {
    this.fetchRegisteredPlayers();
    this.ilutulestikudService.availableColors().subscribe(
      fetchedColorsObject => this.parseColors(fetchedColorsObject),
      thrownError => this.handleError(thrownError),
      () => {});
  }

  ngOnDestroy(): void
  {
    if (this.gameTurnSummariesSubscription)
    {
      this.gameTurnSummariesSubscription.unsubscribe();
    }

    this.selectedPlayer = null;
  }

  dismissErrorMessage(): void
  {
    this.informationText = null;
  }

  fetchRegisteredPlayers(): void
  {
    this.ilutulestikudService.registeredPlayers().subscribe(
      fetchedPlayersObject => this.parsePlayers(fetchedPlayersObject),
      thrownError => this.handleError(thrownError),
      () => {});
  }

  parsePlayers(fetchedPlayersObject: Object): void
  {
    this.registeredPlayers.length = 0;

    // fetchedPlayersObject["Players"] is only an "array-like object", not an array, so does not have foreach.
    for (const fetchedPlayer of fetchedPlayersObject["Players"])
    {
      this.registeredPlayers.push(new Player(fetchedPlayer));
    }
  }

  selectPlayer(selectedPlayer: Player): void
  {
    this.selectedPlayer = selectedPlayer;

    // Once we know the player, we can start checking for games.
    this.startRefreshingGameTurnSummaries();
  }
  
  parseColors(fetchedColorsObject: Object): void
  {
    this.availableColors.length = 0;

    // fetchedColorsObject["Colors"] is only an "array-like object", not an array, so does not have foreach.
    for (const chatColor of fetchedColorsObject["Colors"])
    {
      this.availableColors.push(chatColor);
    }
  }

  changeChatColor(newChatColor: string): void
  {
    this.selectedPlayer.Color = newChatColor;
    this.informationText = null;
    this.ilutulestikudService.updatePlayer(this.selectedPlayer).subscribe(
      returnedPlayersObject => this.parsePlayers(returnedPlayersObject),
      thrownError => this.handleError(thrownError),
      () => {});
  }
  
  openAddPlayerDialog(): void
  {
    let dialogRef = this.materialDialog.open(AddPlayerDialogComponent, {
      width: '250px'
    });

    dialogRef.afterClosed().subscribe(resultFromClose =>
    {
      if (resultFromClose)
      {
        this.informationText = null;
        this.ilutulestikudService.newPlayer(resultFromClose).subscribe(
          returnedPlayersObject => this.parsePlayers(returnedPlayersObject),
          thrownError => this.handleError(thrownError),
          () => {});
      }
    });
  }

  openCreateGameDialog(): void
  {
    // We update the list of available players and also fetch the available rulesets,
    // so that the dialog has enough information to create a valid game.
    this.fetchRegisteredPlayers();

    this.ilutulestikudService.availableRulesets().subscribe(
      fetchedRulesetListObject => this.parseRulesets(fetchedRulesetListObject),
      thrownError => this.handleError(thrownError),
      () => {});
    

    let dialogRef = this.materialDialog.open(CreateGameDialogComponent, {
      width: '250px',
      data: {
        creatingPlayer: this.selectedPlayer,
        availablePlayers: this.registeredPlayers,
        availableRulesets: this.availableRulesets
      }
    });

    dialogRef.afterClosed().subscribe(resultFromClose =>
    {
      if (resultFromClose)
      {
        this.informationText = null;
        this.ilutulestikudService
          .newGame(resultFromClose["GameName"], resultFromClose["RulesetIdentifier"], resultFromClose["PlayerNames"])
          .subscribe(
            () => {},
            thrownError => this.handleError(thrownError),
            () => {});
      }
    });
  }

  parseRulesets(fetchedRulesetListObject: Object): void
  {
    this.availableRulesets.length = 0;

    // fetchedPlayersObject["Players"] is only an "array-like object", not an array, so does not have foreach.
    for (const fetchedRuleset of fetchedRulesetListObject["Rulesets"])
    {
      this.availableRulesets.push(new Ruleset(fetchedRuleset));
    }
  }

  startRefreshingGameTurnSummaries(): void
  {
    this.gameTurnSummariesSubscription =
      Observable
        .timer(0, 1000)
        .takeWhile(() => (this.selectedPlayer != null))
        .subscribe(
          () => this.refreshGameTurnSummaries(),
          thrownError => this.handleError(thrownError),
          () => {});
  }

  refreshGameTurnSummaries(): void
  {
    // We only request new game turn summaries if we are not waiting for the response to the last request.
    if (!this.isAwaitingGameTurnSummaries)
    {
      // We note that we are now awaiting the HTTP response (this.isAwaitingGameTurnSummaries will be
      // set back to false by this.parseGameTurnSummaries(fetchedGameTurnSummaries) which will run when
      // we get the response to the request).
      this.isAwaitingGameTurnSummaries = true;
      this.ilutulestikudService
        .gamesWithPlayer(this.selectedPlayer.ForBackend)
        .subscribe(
          fetchedGameTurnSummaries => this.parseGameTurnSummaries(fetchedGameTurnSummaries),
          thrownError => this.handleError(thrownError),
          () => {});
    }
  }

  parseGameTurnSummaries(fetchedGameTurnSummaries: Object): void
  {
    // If we have received a game turn summary list to parse, we are no longer waiting for the HTTP request to complete.
    this.isAwaitingGameTurnSummaries = false;

    // The object fetchedGameTurnSummaries["TurnSummaries"] is only an "array-like object",
    // not an array, so does not have foreach or length.
    const fetchedSummaries: string[] = Array.from(fetchedGameTurnSummaries["TurnSummaries"]);

    for (var gameIndex: number = 0; gameIndex < fetchedSummaries.length; ++gameIndex)
    {
      const fetchedSummary: Object = fetchedSummaries[gameIndex];

      // First of all we reduce the number of TurnSummary objects if there were more than the refresh returned.
      if (this.turnSummariesOfGamesWithPlayer.length > fetchedSummaries.length)
      {
        this.turnSummariesOfGamesWithPlayer.length = fetchedSummaries.length;
      }

      // We could replace each TurnSummary with each refresh, but that leads to an annoying animated screen refresh.
      // Therefore we update existing TurnSummary objects and only add new ones when necessary.
      if (gameIndex < this.turnSummariesOfGamesWithPlayer.length)
      {
        this.turnSummariesOfGamesWithPlayer[gameIndex].refreshFromSource(fetchedSummary)
      }
      else
      {
        this.turnSummariesOfGamesWithPlayer.push(new TurnSummary(fetchedSummary));
      }
    }

    if (this.turnSummariesOfGamesWithPlayer
      && (this.turnSummariesOfGamesWithPlayer.length > 0)
      && !this.selectedGameIdentification)
    {
      this.selectedGameTabIndex = 0;
      this.selectedGameIdentification = this.turnSummariesOfGamesWithPlayer[this.selectedGameTabIndex].GameForBackend;
    }
  }

  selectGameTab(tabSelection: MatTabChangeEvent): void
  {
    this.selectedGameTabIndex = tabSelection.index;
    this.selectedGameIdentification = this.turnSummariesOfGamesWithPlayer[this.selectedGameTabIndex].GameForBackend;
  }

  handleError(thrownError: Error): void
  {
    console.log(thrownError);

    // We parse the standard "error" out of thrownError, but the backend sends errors as JSON in the form
    // {"Error": "the error message"}, so we parse out "Error" from the parsed-out "error".
    const errorFromError: string = thrownError["error"]
    const errorFromBackend: string = errorFromError["Error"]

    this.informationText = errorFromBackend ? "Backend: " + errorFromBackend : errorFromError;
  }
}
