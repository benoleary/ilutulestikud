import { Component } from '@angular/core';
import { OnInit, OnDestroy } from '@angular/core/src/metadata/lifecycle_hooks';
import { MatDialog } from '@angular/material';
import { Observable } from 'rxjs/Rx';
import { IlutulestikudService } from './ilutulestikud.service';
import { Player } from './models/player.model'
import { TurnSummary } from './models/turnsummary.model'
import { AddPlayerDialogueComponent } from './components/addplayerdialogue.component'
import { CreateGameDialogueComponent } from './components/creategamedialogue.component'
import { Subscription } from 'rxjs/Subscription';

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
  namesOfGamesWithPlayer: string[];
  gameNamesTimerSubscription: Subscription;
  gameTurnSummariesSubscription: Subscription;
  isAwaitingGameTurnSummaries: boolean;
  

  constructor(public ilutulestikudService: IlutulestikudService, public materialDialog: MatDialog)
  {
    this.selectedPlayer = null;
    this.registeredPlayers = [];
    this.availableColors = [];
    this.informationText = null;
    this.namesOfGamesWithPlayer = [];
    this.gameNamesTimerSubscription = null;
    this.gameTurnSummariesSubscription = null;
    this.isAwaitingGameTurnSummaries = false;
  }

  ngOnInit(): void
  {
    // We do not start this.refreshGames(); in ngOnInit() as it somehow leads to the display not ever updating.
    this.ilutulestikudService.registeredPlayers().subscribe(
      fetchedPlayersObject => this.parsePlayers(fetchedPlayersObject),
      thrownError => this.handleError(thrownError),
      () => {});
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

    if (this.gameNamesTimerSubscription)
    {
      this.gameNamesTimerSubscription.unsubscribe();
    }

    this.selectedPlayer = null;
  }

  dismissErrorMessage(): void
  {
    this.informationText = null;
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
    this.refreshGames();
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
    let dialogRef = this.materialDialog.open(AddPlayerDialogueComponent, {
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
    let dialogRef = this.materialDialog.open(CreateGameDialogueComponent, {
      width: '250px',
      data: {
        creatingPlayer: this.selectedPlayer.Name,
        availablePlayers: this.registeredPlayers.map(registerPlayer => registerPlayer.Name)
      }
    });

    dialogRef.afterClosed().subscribe(resultFromClose =>
    {
      if (resultFromClose)
      {
        this.informationText = null;
        this.ilutulestikudService
          .newGame(resultFromClose["Name"], resultFromClose["Players"])
          .subscribe(
            () => {},
            thrownError => this.handleError(thrownError),
            () => {});
      }
    });
  }

  startRefreshingGameTurnSummaries(): void
  {
    this.gameTurnSummariesSubscription =
      Observable
        .timer(2000,1000)
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
        .gamesWithPlayer(this.selectedPlayer.Name)
        .subscribe(
          fetchedGameTurnSummaries => this.parseGameTurnSummaries(fetchedGameTurnSummaries),
          thrownError => this.handleError(thrownError),
          () => {});
    }
  }

  refreshGames(): void
  {
    // Once the result has been parsed, this.parseGamesAndRefresh(...) will trigger another
    // call of refreshGames() after a second.
    this.ilutulestikudService.gamesWithPlayer(this.selectedPlayer.Name).subscribe(
        fetchedGamesObject => this.parseGamesAndRefresh(fetchedGamesObject),
        thrownError => this.handleError(thrownError),
        () => {});
  }

  parseGameTurnSummaries(fetchedGameTurnSummaries: Object): void
  {
    // If we have received a game turn summary list to parse, we are no longer waiting for the HTTP request to complete.
    this.isAwaitingGameTurnSummaries = false;

    // Will re-instate when satisfied that this can replace parseGamesAndRefresh(...) etc.
    // this.namesOfGamesWithPlayer.length = 0;

    // fetchedGamesObject["Games"] is only an "array-like object", not an array, so does not have foreach.
    for (const turnSummary of fetchedGameTurnSummaries["TurnSummaries"])
    {
      console.log("turnSummary = " + JSON.stringify(turnSummary));
      // Will re-instate when satisfied that this can replace parseGamesAndRefresh(...) etc.
      // this.namesOfGamesWithPlayer.push(turnSummary["GameName"]);

      // Plus some more display...
    }
  }

  parseGamesAndRefresh(fetchedGamesObject: Object): void
  {
    this.namesOfGamesWithPlayer.length = 0;

    // fetchedGamesObject["Games"] is only an "array-like object", not an array, so does not have foreach.
    for (const gameName of fetchedGamesObject["Games"])
    {
      this.namesOfGamesWithPlayer.push(gameName);
    }

    // Once the result has been parsed, we wait a second, then trigger another
    // call of refreshGames().
    this.gameNamesTimerSubscription = Observable.timer(1000).first().subscribe(
      () => this.refreshGames());
  }

  handleError(thrownError: Error): void
  {
    console.log("Error! " + JSON.stringify(thrownError));
    this.informationText = "Error! " + JSON.stringify(thrownError["error"]);
  }
}
