package variants

import (
	"fmt"
	"html"
	"strings"
)

// SmallGods is variant for the Jezero game inspired by Discworld
type SmallGods struct{}

// NewSmallGods creates a new instance of SmallGods
func NewSmallGods() *SmallGods { return &SmallGods{} }

////////////////////////////////////////////////////////////////////////////////
//revive:disable:exported

func (h *SmallGods) ViewTemplateName() string    { return "viewCarcassonne" }
func (h *SmallGods) TemplateStateName() string   { return "Nálada města" }
func (h *SmallGods) TemplateStateSymbol() string { return "☺" }
func (h *SmallGods) TemplateMoneyName() string   { return "Božské body" }
func (h *SmallGods) TemplateMoneySymbol() string { return "" }

func (h *SmallGods) NopName() string        { return "Nedělat nic" }
func (h *SmallGods) EcoName() string        { return "Poklidné kázání na rohu" }
func (h *SmallGods) HarvestName() string    { return "Únosy nevěřících do kláštera" }
func (h *SmallGods) CleaningName() string   { return "Přispět městu na chléb a hry" }
func (h *SmallGods) InspectionName() string { return "Bonzovat patriciovi" }
func (h *SmallGods) EspionageName() string  { return "Všeobecná zpověď" }

////////////////////////////////////////////////////////////////////////////////

func (h *SmallGods) EcoMessage(money int, pollution int) string {
	return fmt.Sprintf("Kázali jste poklidně na rohu. Získali jste tím pro svého boha %d božských bodů a vaše město přišlo o %d☺.", money, pollution)
}
func (h *SmallGods) HarvestPenaltyMessage(penalty int) string {
	return fmt.Sprintf("Vaše únosy do kláštera byly odhaleny patricijem! To zostudilo vašeho boha, ztratili jste %d božských bodů.", penalty)
}
func (h *SmallGods) HarvestSuccessMessage(money int, pollution int) string {
	return fmt.Sprintf("Unesli jste několik nevěřících do kláštera a důsledným kázáním jste je obrátili na svoji víru. Spokojenost města klesla o %d☺, ale získali jste %d božských bodů.", pollution, money)
}
func (h *SmallGods) CleaningMessage(cleaning int) string {
	return fmt.Sprintf("Přispěli jste městu na chléb a hry, městu se zvýšila spokojenost o %d☺.", cleaning)
}
func (h *SmallGods) InspectionMessage() string {
	return "Požádali jste patricije o kontrolu ve městě. Pokud někdo v minulém kole prováděl něco špatného, tak byl potrestán."
}
func (h *SmallGods) EspionageFailMessage() string {
	return "Všeobecnou zpověď vám bohužel překazilo patriciova inspekce, nic jste nezjistili."
}
func (h *SmallGods) EspionageSuccessMessage(teamActions map[string]string) string {
	results := []string{}
	for team, action := range teamActions {
		results = append(results, fmt.Sprintf("<li>%s: <b>%s</b></li>", html.EscapeString(team), action))
	}
	return fmt.Sprintf("Úspěšně jste uspořádali veřejné zpovědi. Akce v proběhlém kole:<ul>\n%s\n</ul>", strings.Join(results, "\n"))
}

func (h *SmallGods) GlobalMessage(reduce string, increase string, change int) string {
	return fmt.Sprintf("<b>Stěhování:</b> Stěhováním se město nálada města %s zmenšila o %d☺ a nálada města %s se naopak zvedla o %d☺.", reduce, change, increase, change)
}
